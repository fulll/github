#!/usr/bin/env bash
PROJECT_NAME="github"

: ${USE_SUDO:="true"}
: ${GH_CLI_INSTALL_DIR:="/usr/local/bin"}


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

# initOS discovers the operating system for this system.
initCLIPath() {
  suffix=""

  if [ "${OS}" = "windows" ]; then
    suffix=".exe"
  fi  

  GH_CLI_TMP="libs/${OS}-${ARCH}/github${suffix}"
  echo "Init CLI Path for ${OS}-${ARCH} (${GH_CLI_TMP})"
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
  echo "check support ${GH_CLI_TMP}"
  if [ ! -f "${GH_CLI_TMP}" ]; then
    echo "No prebuilt binary for ${OS}-${ARCH}."
    echo "To build from source, go to https://github.com/fulll/github"
    exit 1   
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


# installFile verifies the SHA256 for the file, then unpacks and
# installs it.
installFile() {
  echo "Preparing to install $PROJECT_NAME from ${GH_CLI_TMP} ${GH_CLI_INSTALL_DIR}"
  chmod +x "${GH_CLI_TMP}"
  runAsRoot cp "${GH_CLI_TMP}" "$GH_CLI_INSTALL_DIR/$PROJECT_NAME"
  ls $GH_CLI_INSTALL_DIR
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
    echo -e "\tFor support, go to https://github.com/fulll/github."
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

# cleanup temporary files to avoid https://github.com/fulll/github/issues/2977
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
initCLIPath
verifySupported
installFile
testVersion
