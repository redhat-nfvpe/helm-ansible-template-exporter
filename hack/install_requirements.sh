#!/usr/bin/env bash
set -e
install_operator_sdk() {
  if [[ ! -z "${INSTALL_OPERATOR_SDK}" ]]; then
    # Install operator-sdk
    if ! which operator-sdk 2>&1 >/dev/null; then
      sudo wget https://github.com/operator-framework/operator-sdk/releases/download/v0.17.0/operator-sdk-v0.17.0-x86_64-linux-gnu -O /usr/local/bin/operator-sdk
      sudo chmod 755 /usr/local/bin/operator-sdk
    fi
  else
    echo "Did not find operator-sdk, set INSTALL_OPERATOR_SDK=1 in env.sh or set via cli,and try again to install operator-sdk"
    exit 1
  fi
}
