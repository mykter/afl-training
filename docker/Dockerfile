FROM ubuntu:18.04
LABEL maintainer="Michael Macnair"

RUN echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections

# SSHd
#############

RUN apt-get update && apt-get install -y openssh-server
RUN mkdir /var/run/sshd

# SSH login fix. Otherwise user is kicked off after login
RUN sed 's@session\s*required\s*pam_loginuid.so@session optional pam_loginuid.so@g' -i /etc/pam.d/sshd

ENV NOTVISIBLE "in users profile"
RUN echo "export VISIBLE=now" >> /etc/profile

# Users
##############
# Root's password is root, but we don't need to enable remote login as root
RUN echo 'root:root' | chpasswd
#RUN sed -i 's/PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config

RUN useradd --create-home --shell /bin/bash fuzzer
# See the README - this password is visible to anyone with access to the image
COPY password.txt .
RUN echo "fuzzer:`cat password.txt`" | chpasswd && rm password.txt

# AFL + Deps
##############
USER root
RUN apt-get update && apt-get install -y clang-6.0 llvm-6.0-dev git build-essential curl emacs vim nano libssl-dev screen cgroup-bin sudo
RUN update-alternatives --install /usr/bin/clang clang `which clang-6.0` 1
RUN update-alternatives --install /usr/bin/clang++ clang++ `which clang++-6.0` 1
RUN update-alternatives --install /usr/bin/llvm-config llvm-config `which llvm-config-6.0` 1
RUN update-alternatives --install /usr/bin/llvm-symbolizer llvm-symbolizer `which llvm-symbolizer-6.0` 1

# For the date exercise:
RUN apt-get install -y autopoint bison gperf autoconf texinfo gettext

# For sudo for ASAN:
RUN usermod -aG sudo fuzzer

ENV AFLVERSION=afl-2.52b
USER fuzzer
WORKDIR /home/fuzzer
RUN curl http://lcamtuf.coredump.cx/afl/releases/$AFLVERSION.tgz | tar xz
RUN git clone https://github.com/vanhauser-thc/afl-patches.git

WORKDIR /home/fuzzer/$AFLVERSION
RUN patch -p0 <../afl-patches/afl-llvm-fix.diff
RUN patch -p0 <../afl-patches/afl-llvm-fix2.diff
RUN patch -p0 <../afl-patches/afl-sort-all_uniq-fix.diff
RUN make && cd llvm_mode && make

# It's important that AFL comes last in the PATH, as otherwise its `as` overrides the system one
RUN echo PATH=$PATH:$HOME/$AFLVERSION >> ~/.bashrc
RUN echo AFL_PATH=$HOME/$AFLVERSION/ >> ~/.bashrc

# You could install gnuplot-nox, but it increases the image size a lot (~100 extra packages).
# Students can install it themselves if they want it.

# Exercises
##############
USER fuzzer
WORKDIR /home/fuzzer
RUN git clone https://github.com/ThalesIgnite/afl-training.git workshop
# Use this if building using a local copy of the training materials
# ADD afl-training ./afl-training
# USER root
# RUN chown -R fuzzer:fuzzer afl-training

# By default just run an SSH daemon. To run locally instead, use something like this:
#    docker run -it --user fuzzer afl-training:latest /bin/bash
##############
USER root
EXPOSE 22
CMD ["/usr/sbin/sshd", "-D"]
