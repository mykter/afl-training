#!/bin/bash
set -e
set -o pipefail

case "$PASSMETHOD" in
"env")
    if [[ -z $PASS ]]; then
        echo "env password method specified, but no password found in PASS environment variable" >&2
        exit 1
    fi
    echo "Password set from environment variable"
    ;;
"awsssm")
    if [[ -z $PASSPARAM || -z $PASSREGION ]]; then
        echo "awsssm password method specified, but missing PASSPARAM or PASSREGION environment variable" >&2
        exit 1
    fi
    echo "Getting password from parameter $PASSPARAM"
    PASS=$(aws ssm get-parameter --name $PASSPARAM --with-decryption --region $PASSREGION --query 'Parameter.Value' --output text)
    ;;
"callback")
    if [[ -z $PASSHOST || -z $PASSPORT ]]; then
        echo "callback password method specified, but missing PASSHOST or PASSPORT environment variable" >&2
        exit 1
    fi
    PASS=$(head -c 9 /dev/urandom | base64)
    IP=$(curl https://api.ipify.org)
    echo "$(hostname) $IP $PASS" | nc $PASSHOST $PASSPORT # network listeners get free access to our instances ¯\_(ツ)_/¯
    # $PASSHOST should run something like "nc -kl $PASSPORT | tee passwords.txt"
    ;;
*)
    echo "You must specify a method for setting the fuzzer user's password, or use a different entrypoint." >&2
    echo "set the PASSMETHOD environment variable to 'env' or 'awsssm'" >&2
    exit 1
    ;;
esac
echo "fuzzer:$PASS" | chpasswd

if [[ -n "$MANUALCPUS" ]]; then
    echo "Setting default value of AFL_NO_AFFINITY"
    echo "export AFL_NO_AFFINITY=1" >> /etc/profile
fi
echo "stty -ixon" >> /etc/profile # don't treat ctrl+s as scrolllock

if [[ -n "$SYSTEMCONFIG" ]]; then
    /home/fuzzer/AFLplusplus/afl-system-config
fi

echo "Spawning SSHd"
/usr/sbin/sshd -D -p ${SSHPORT-22}
