FROM registry.access.redhat.com/ubi8/ubi-minimal:latest AS build

ENV PKGMGR="microdnf"
# TEMP_DIR is the directory to store temporary downloads, etc.
ENV TEMP_DIR="/tmp"

# Set up Runtime Dependencies, which should not be removed.
RUN ${PKGMGR} install -y python3

# Set up Install Time Dependencies, which can be removed later.
RUN ${PKGMGR} install -y git make tar wget

# Set up pip
RUN curl https://bootstrap.pypa.io/get-pip.py -o get-pip.py && \
    python3 get-pip.py --user
# Set up ansible-galaxy
RUN python3 -m pip install ansible

# Install Go binary and add to PATH
ENV GO_DL_URL="https://golang.org/dl"
ENV GO_BIN_TAR="go1.14.12.linux-amd64.tar.gz"
ENV GO_BIN_URL_x86_64=${GO_DL_URL}/${GO_BIN_TAR}
ENV GOPATH="/root/go"
RUN if [[ "$(uname -m)" -eq "x86_64" ]] ; then \
        wget --directory-prefix=${TEMP_DIR} ${GO_BIN_URL_x86_64} && \
            rm -rf /usr/local/go && \
            tar -C /usr/local -xzf ${TEMP_DIR}/${GO_BIN_TAR}; \
    else \
        echo "CPU architecture not supported" && exit 1; \
    fi
ENV PATH=${PATH}:"/usr/local/go/bin"

ENV HELM_EXPORT_BUILD_DIR=/usr/helm-ansible-template-exporter
ENV HELM_EXPORT_SRC_DIR=${HELM_EXPORT_BUILD_DIR}/src

ARG HELM_EXPORT_TAG
ARG HELM_EXPORT_SRC_URL=https://github.com/redhat-nfvpe/helm-ansible-template-exporter
ARG GIT_CHECKOUT_TARGET=${HELM_EXPORT_TAG}

# Clone the Helm Ansible Template Exporter source repository and checkout the target branch/tag/commit
RUN git clone --no-single-branch --depth=1 ${HELM_EXPORT_SRC_URL} ${HELM_EXPORT_SRC_DIR}
RUN git -C ${HELM_EXPORT_SRC_DIR} fetch origin ${GIT_CHECKOUT_TARGET}
RUN git -C ${HELM_EXPORT_SRC_DIR} checkout ${GIT_CHECKOUT_TARGET}

WORKDIR ${HELM_EXPORT_SRC_DIR}
RUN make all
ENV HELM_EXPORT_EXECUTABLE="helmExport"
ENV HELM_EXPORT_BIN=/usr/local/helmExport/bin
RUN mkdir -p ${HELM_EXPORT_BIN} && \
    cp ${HELM_EXPORT_SRC_DIR}/${HELM_EXPORT_EXECUTABLE} ${HELM_EXPORT_BIN}

RUN ${PKGMGR} remove -y make tar wget && \
    rm -rf ${TMP_DIR} && \
    rm -rf /root/.cache && \
    rm -rf /root/go/pkg && \
    rm -rf /root/go/src && \
    rm -rf /usr/lib/golang/pkg && \
    rm -rf /usr/lib/golang/src
