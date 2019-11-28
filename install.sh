#!/usr/bin/env bash
PROJECT_NAME="github"

: ${USE_SUDO:="true"}
: ${GH_CLI_INSTALL_DIR:="/usr/local/bin"}

CURL_OPTS="-SsL"
if ! [[ -z "$GITHUB_USER" && -z "$GITHUB_TOKEN" ]] ; then
  CURL_OPTS="$CURL_OPTS -u $GITHUB_USER:$GITHUB_TOKEN"
fi

# initArch discovers the architecture for this system.
initArch() {
  ARCH=$(uname -m)
  case $ARCH in
    armv5*) ARCH="armv5";;
    armv6*) ARCH="armv6";;
    armv7*) ARCH="arm";;
    aarch64) ARCH="arm64";;
    x86) ARCH="386";;
    x86_64) ARCH="amd64";;
    i686) ARCH="386";;
    i386) ARCH="386";;
  esac
}

# initOS discovers the operating system for this system.
initOS() {
  OS=$(echo `uname`|tr '[:upper:]' '[:lower:]')

  case "$OS" in
    # Minimalist GNU for Windows
    mingw*) OS='windows';;
  esac
}

# runs the given command as root (detects if we are root already)
runAsRoot() {
  local CMD="$*"

  if [ $EUID -ne 0 -a $USE_SUDO = "true" ]; then
    CMD="sudo $CMD"
  fi

  $CMD
}

# verifySupported checks that the os/arch combination is supported for
# binary builds.
verifySupported() {
  local supported="darwin-amd64\nlinux-amd64\nwindows-amd64"
  if ! echo "${supported}" | grep -q "${OS}-${ARCH}"; then
    echo "No prebuilt binary for ${OS}-${ARCH}."
    echo "To build from source, go to https://github.com/inextensodigital/github"
    exit 1
  fi

  if ! type "curl" > /dev/null ; then
    echo "curl is required"
    exit 1
  fi

  if ! type "jq" > /dev/null ; then
    echo "jq is required"
    exit 1
  fi
}

# checkDesiredVersion checks if the desired version is available.
checkDesiredVersion() {
  if [ "x$DESIRED_VERSION" == "x" ]; then
    # Get tag from release URL
    local latest_release_url="https://github.com/inextensodigital/github/releases/latest"
    TAG=$(curl $CURL_OPTS -o /dev/null -w %{url_effective} $latest_release_url | grep -oE "[^/]+$" )
  else
    TAG=$DESIRED_VERSION
  fi
}

# checkGithubInstalledVersion checks which version of github is installed and
# if it needs to be changed.
checkGithubInstalledVersion() {
  if [[ -f "${GH_CLI_INSTALL_DIR}/${PROJECT_NAME}" ]]; then
    local version=$(github version | grep '^github version ' | cut -d' ' -f3)
    if [[ "$version" == "$TAG" ]]; then
      echo "github ${version} is already ${DESIRED_VERSION:-latest}"
      return 0
    else
      echo "github ${TAG} is available. Changing from version ${version}."
      return 1
    fi
  else
    return 1
  fi
}

# downloadFile downloads the latest binary package and also the checksum
# for that binary.
downloadFile() {
  DOWNLOAD_URL=$(
    curl $CURL_OPTS https://api.github.com/repos/inextensodigital/github/releases/tags/$TAG |
    jq -r '.assets[] | .browser_download_url' | grep -E "$OS-$ARCH(.exe)?\$"
  )
  CHECKSUM_URL=$(
    curl $CURL_OPTS https://api.github.com/repos/inextensodigital/github/releases/tags/$TAG |
    jq -r '.assets[] | .browser_download_url' | grep -E "$OS-$ARCH(.exe)?\$"
  )
  CHECKSUM_URL="$DOWNLOAD_URL.sha256"
  GH_CLI_TMP_ROOT="$(mktemp -dt github-installer-XXXXXX)"
  GH_CLI_TMP_FILE="$GH_CLI_TMP_ROOT/$PROJECT_NAME"
  GH_CLI_SUM_FILE="$GH_CLI_TMP_ROOT/$PROJECT_NAME.sha256"
  echo "Downloading $DOWNLOAD_URL"
  curl $CURL_OPTS "$CHECKSUM_URL" -o "$GH_CLI_SUM_FILE"
  curl $CURL_OPTS "$DOWNLOAD_URL" -o "$GH_CLI_TMP_FILE"
}

# installFile verifies the SHA256 for the file, then unpacks and
# installs it.
installFile() {
  GH_CLI_TMP="$GH_CLI_TMP_ROOT/$PROJECT_NAME"
  local sum=$(openssl sha1 -sha256 ${GH_CLI_TMP_FILE} | awk '{print $2}')
  local expected_sum=$(cat ${GH_CLI_SUM_FILE} | awk '{print $1}')
  if [ "$sum" != "$expected_sum" ]; then
    echo "SHA sum of ${GH_CLI_TMP_FILE} does not match. Aborting."
    exit 1
  fi

  echo "Preparing to install $PROJECT_NAME into ${GH_CLI_INSTALL_DIR}"
  chmod +x "$GH_CLI_TMP"
  runAsRoot cp "$GH_CLI_TMP" "$GH_CLI_INSTALL_DIR"
  echo "$PROJECT_NAME installed into $GH_CLI_INSTALL_DIR/$PROJECT_NAME"
}

# fail_trap is executed if an error occurs.
fail_trap() {
  result=$?
  if [ "$result" != "0" ]; then
    if [[ -n "$INPUT_ARGUMENTS" ]]; then
      echo "Failed to install $PROJECT_NAME with the arguments provided: $INPUT_ARGUMENTS"
      help
    else
      echo "Failed to install $PROJECT_NAME"
    fi
    echo -e "\tFor support, go to https://github.com/inextensodigital/github."
  fi
# cleanup
  exit $result
}

# testVersion tests the installed client to make sure it is working.
testVersion() {
  set +e
  GITHUB="$(which $PROJECT_NAME)"
  if [ "$?" = "1" ]; then
    echo "$PROJECT_NAME not found. Is $GH_CLI_INSTALL_DIR on your "'$PATH?'
    exit 1
  fi
  set -e
}

# help provides possible cli installation arguments
help () {
  echo "Accepted cli arguments are:"
  echo -e "\t[--help|-h ] ->> prints this help"
  echo -e "\t[--version|-v <desired_version>] . When not defined it defaults to latest"
  echo -e "\te.g. --version v2.4.0  or -v latest"
  echo -e "\t[--no-sudo]  ->> install without sudo"
}

# cleanup temporary files to avoid https://github.com/inextensodigital/github/issues/2977
cleanup() {
  if [[ -d "${GH_CLI_TMP_ROOT:-}" ]]; then
    rm -rf "$GH_CLI_TMP_ROOT"
  fi
}

# Execution

#Stop execution on any error
trap "fail_trap" EXIT
set -e

# Parsing input arguments (if any)
export INPUT_ARGUMENTS="${@}"
set -u
while [[ $# -gt 0 ]]; do
  case $1 in
    '--version'|-v)
       shift
       if [[ $# -ne 0 ]]; then
           export DESIRED_VERSION="${1}"
       else
           echo -e "Please provide the desired version. e.g. --version v2.4.0 or -v latest"
           exit 0
       fi
       ;;
    '--no-sudo')
       USE_SUDO="false"
       ;;
    '--help'|-h)
       help
       exit 0
       ;;
    *) exit 1
       ;;
  esac
  shift
done
set +u

initArch
initOS
verifySupported
checkDesiredVersion
if ! checkGithubInstalledVersion; then
  downloadFile
  installFile
fi
testVersion
cleanup
