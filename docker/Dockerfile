FROM ubuntu:16.04
MAINTAINER Michael Macnair

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
RUN apt-get update && apt-get install -y clang llvm-dev git build-essential curl vim nano libssl-dev screen cgroup-bin sudo
RUN update-alternatives --install /usr/bin/clang clang `which clang-3.8` 1
RUN update-alternatives --install /usr/bin/clang++ clang++ `which clang++-3.8` 1
RUN update-alternatives --install /usr/bin/llvm-config llvm-config `which llvm-config-3.8` 1
RUN update-alternatives --install /usr/bin/llvm-symbolizer llvm-symbolizer `which llvm-symbolizer-3.8` 1

USER fuzzer
WORKDIR /home/fuzzer
RUN curl http://lcamtuf.coredump.cx/afl/releases/afl-latest.tgz | tar xz

# You could install gnuplot-nox, but it increases the image size a lot (~100 extra packages).
# Students can install it themselves if they want it.

# Exercises
##############
USER fuzzer
WORKDIR /home/fuzzer
RUN git clone https://github.com/ThalesIgnite/afl-training.git
# Use this if building using a local copy of the training materials
# ADD afl-training ./afl-training
# USER root
# RUN chown -R fuzzer:fuzzer afl-training

# sshd mode
###############
USER root
EXPOSE 22
CMD ["/usr/sbin/sshd", "-D"]
