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
"gcpmeta")
    PASS=$(head -c 9 /dev/urandom | base64)
    echo $PASS | http --check-status PUT http://metadata.google.internal/computeMetadata/v1/instance/guest-attributes/fuzzing/password Metadata-Flavor:Google
    ;;
*)
    echo "You must specify a method for setting the fuzzer user's password, or use a different entrypoint." >&2
    echo "set the PASSMETHOD environment variable to 'env', 'awsssm', 'gcpmeta', or 'callback'" >&2
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
    set +e # some aspects of this will often fail; best-efforts is fine for a learning environment
    /home/fuzzer/AFLplusplus/afl-system-config
fi

echo "Spawning SSHd on port ${SSHPORT-2222}"
/usr/sbin/sshd -D -p ${SSHPORT-2222}
