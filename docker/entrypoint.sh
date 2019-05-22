#!/bin/bash
set -e
set -o pipefail

if [[ "$PASSMETHOD" == "env" ]]; then
    if [[ -z $PASS ]]; then
        echo "env password method specified, but no password found in PASS environment variable" >&2
        exit 1
    fi
    echo "Password set from environment variable"
elif [[ "$PASSMETHOD" == "awsssm" ]]; then
    if [[ -z $PASSPARAM -o -z $PASSREGION ]]; then
        echo "awsssm password method specified, but missing PASSPARAM or PASSREGION environment variable" >&2
        exit 1
    fi
    echo "Getting password from parameter $PASSPARAM"
    PASS=$(aws ssm get-parameter --name $PASSPARAM --with-decryption --region $PASSREGION | jq -r '.Parameter.Value')
else
    echo "You must specify a method for setting the fuzzer user's password, or use a different entrypoint." >&2
    echo "set the PASSMETHOD environment variable to 'env' or 'awsssm'" >&2
    exit 1
fi

echo "fuzzer:$PASS" | chpasswd

echo "Spawning SSHd"
/usr/sbin/sshd -D