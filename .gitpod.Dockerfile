FROM gitpod/workspace-full

# Install Google Cloud SDK
RUN curl -sSL https://sdk.cloud.google.com | bash

# Install Go (if not already present)
# RUN wget https://go.dev/dl/go1.22.3.linux-amd64.tar.gz && \
#     tar -C /usr/local -xzf go1.22.3.linux-amd64.tar.gz && \
#     rm go1.22.3.linux-amd64.tar.gz

ENV PATH="/workspace/google-cloud-sdk/bin:/usr/local/go/bin:${PATH}"