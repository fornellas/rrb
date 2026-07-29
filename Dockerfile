FROM ubuntu:26.04
RUN apt-get update && apt-get -y --no-install-recommends install \
    curl \
    build-essential \
    ca-certificates \
    gawk \
    git \
    less \
    unzip
RUN passwd -d root
ARG USER
ARG GROUP
ARG UID
ARG GID
ARG HOME
RUN if id --user ${UID} ; then userdel $(awk -F: -v uid=${UID} '$3==uid{print $1}' < /etc/passwd) ; fi
RUN if gawk -F: -v gid=${GID} '$3==gid{found=1}END{if(found!=1){exit 1}}' < /etc/group ; then groupdel ${GID} ; fi
RUN groupadd --gid ${GID} ${GROUP}
RUN mkdir ${HOME}
RUN useradd --home-dir ${HOME} --gid ${GID} --no-create-home --shell /bin/bash --uid ${UID} ${USER}
RUN chown ${UID}:${GID} ${HOME}
# This is required on Darwin, as the UID/GID mappping is not 1:1.
RUN git config --global --add safe.directory ${HOME}/rrb
