FROM ubuntu:20.04
LABEL maintainer="Michael Macnair"

RUN echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections

# We want manpages in the container - the base image excludes them
RUN rm /etc/dpkg/dpkg.cfg.d/excludes

# llvm-11
RUN apt-get update && apt-get install -y --no-install-recommends wget ca-certificates gnupg2 && rm -rf /var/lib/apt/lists
RUN echo deb http://apt.llvm.org/focal/ llvm-toolchain-focal-11 main >> /etc/apt/sources.list
RUN wget -O - https://apt.llvm.org/llvm-snapshot.gpg.key | apt-key add - 

# Packages
##############
# by line:
#   build and afl
#   llvm-11 (for afl-clang-lto)
#   date challenge
#   libxml2 challenge
#   server/entrypoint
#   user tools
#   debugging tools
RUN apt-get update && apt-get install -y \
    git build-essential curl libssl-dev sudo libtool libtool-bin libglib2.0-dev bison flex automake python3 python3-dev python3-setuptools python-is-python3 libpixman-1-dev gcc-9-plugin-dev cgroup-tools \
    clang-11 clang-tools-11 libc++1-11 libc++-11-dev libc++abi1-11 libc++abi-11-dev libclang1-11 libclang-11-dev libclang-common-11-dev libclang-cpp11 libclang-cpp11-dev liblld-11 liblld-11-dev liblldb-11 liblldb-11-dev libllvm11 libomp-11-dev libomp5-11 lld-11 lldb-11 python3-lldb-11 llvm-11 llvm-11-dev llvm-11-runtime llvm-11-tools \
    rsync autopoint bison gperf autoconf texinfo gettext \
    libtool pkg-config libz-dev python2.7-dev \
    awscli openssh-server ncat \
    emacs vim nano screen htop man manpages-posix-dev wget httpie bash-completion ripgrep \
    gdb byobu \
    && rm -rf /var/lib/apt/lists

RUN echo y | unminimize

# Users & SSH
##############
RUN useradd --create-home --shell /bin/bash fuzzer
# See the README - the password is set by the entry script

# passwordless sudo access for ASAN and installing extra tools:
RUN echo "fuzzer ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers
# RUN usermod -aG sudo fuzzer

RUN mkdir /var/run/sshd

# AFL
###########
RUN update-alternatives --install /usr/bin/clang clang $(which clang-11) 1 && \
    update-alternatives --install /usr/bin/clang++ clang++ $(which clang++-11) 1 && \
    update-alternatives --install /usr/bin/llvm-config llvm-config $(which llvm-config-11) 1 && \
    update-alternatives --install /usr/bin/llvm-symbolizer llvm-symbolizer $(which llvm-symbolizer-11) 1 && \
    update-alternatives --install /usr/bin/llvm-cov llvm-cov $(which llvm-cov-11) 1 && \
    update-alternatives --install /usr/bin/llvm-profdata llvm-profdata $(which llvm-profdata-11) 1

# (environment variables won't be visible in the SSH session unless added to /etc/profile or similar)
ENV AFLVERSION=3.0c
USER fuzzer
WORKDIR /home/fuzzer
RUN git clone https://github.com/AFLplusplus/AFLplusplus
WORKDIR /home/fuzzer/AFLplusplus
RUN git checkout $AFLVERSION
RUN make distrib
RUN sudo make install

# You could install gnuplot-nox, but it increases the image size a lot (~100 extra packages).
# Students can install it themselves if they want it.

# Exercises
##############
USER fuzzer
WORKDIR /home/fuzzer
RUN git clone https://github.com/mykter/afl-training.git workshop
# Use this if building using a local copy of the training materials
# ADD . ./local-workshop
# USER root
# RUN chown -R fuzzer:fuzzer local-workshop

WORKDIR /home/fuzzer/workshop
RUN http https://www.eecis.udel.edu/~ntp/ntp_spool/ntp4/ntp-4.2/ntp-4.2.2.tar.gz | tar xz -C challenges/ntpq
RUN http https://www.eecis.udel.edu/~ntp/ntp_spool/ntp4/ntp-4.2/ntp-4.2.8p10.tar.gz | tar xz -C challenges/ntpq
RUN git submodule init && git submodule update

# By default run an SSH daemon. To run locally instead, use something like this:
#    docker run -it --user fuzzer afl-training:latest /bin/bash
##############
USER root

COPY entrypoint.sh /usr/local/bin/entrypoint.sh

EXPOSE 22
CMD ["entrypoint.sh"]
# on some systems you might want to run AFLplusplus/afl-system-config; only works if the container is run in privileged mode and you don't care about security of the host
