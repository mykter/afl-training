FROM ubuntu:18.04
LABEL maintainer="Michael Macnair"

RUN echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections

# We want manpages in the container - the base image excludes them
RUN rm /etc/dpkg/dpkg.cfg.d/excludes

# Packages
##############
# by line:
#   build and afl
#   date challenge
#   libxml2 challenge
#   server/entrypoint
#   user tools
#   debugging tools
USER root
RUN apt-get update && apt-get install -y \
    clang-6.0 llvm-6.0-dev git build-essential curl libssl-dev cgroup-bin sudo \
    rsync autopoint bison gperf autoconf texinfo gettext \
    libtool pkg-config libz-dev python2.7-dev \
    awscli openssh-server \
    emacs vim nano screen htop man manpages-posix-dev \
    lldb gdb

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
RUN update-alternatives --install /usr/bin/clang clang `which clang-6.0` 1 && \
    update-alternatives --install /usr/bin/clang++ clang++ `which clang++-6.0` 1 && \
    update-alternatives --install /usr/bin/llvm-config llvm-config `which llvm-config-6.0` 1 && \
    update-alternatives --install /usr/bin/llvm-symbolizer llvm-symbolizer `which llvm-symbolizer-6.0` 1

# (environment variables won't be visible in the SSH session unless added to /etc/profile or similar)
ENV AFLVERSION=afl-2.52b
USER fuzzer
WORKDIR /home/fuzzer
RUN curl http://lcamtuf.coredump.cx/afl/releases/$AFLVERSION.tgz | tar xz
WORKDIR /home/fuzzer/$AFLVERSION
RUN git clone https://github.com/vanhauser-thc/afl-patches.git && \
    patch -p0 <./afl-patches/afl-llvm-fix.diff && \
    patch -p0 <./afl-patches/afl-llvm-fix2.diff && \
    patch -p0 <./afl-patches/afl-sort-all_uniq-fix.diff
RUN make && cd llvm_mode && make
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

# By default run an SSH daemon. To run locally instead, use something like this:
#    docker run -it --user fuzzer afl-training:latest /bin/bash
##############
USER root

COPY entrypoint.sh /usr/local/bin/entrypoint.sh

EXPOSE 22
CMD ["entrypoint.sh"]
