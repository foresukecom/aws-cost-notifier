# ベースイメージから継承
FROM devcontainer-base:latest

# rootユーザーに切り替えてパッケージインストール
USER root

# Goのインストール（マルチアーキテクチャ対応）
ENV GO_VERSION=1.23.4
RUN ARCH=$(uname -m | sed 's/x86_64/amd64/g' | sed 's/aarch64/arm64/g') \
    && wget -O go.tar.gz "https://go.dev/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz" \
    && tar -C /usr/local -xzf go.tar.gz \
    && rm go.tar.gz

# Go関連の環境変数設定
ENV PATH=/usr/local/go/bin:$PATH
ENV GOPATH=/home/developer/go
ENV GOBIN=$GOPATH/bin
ENV PATH=$GOBIN:$PATH

# 開発者ユーザーに戻る
USER developer

# Goワークスペースディレクトリ作成
RUN mkdir -p $GOPATH/src $GOPATH/bin $GOPATH/pkg

# よく使用されるGoツールのインストール
RUN go install golang.org/x/tools/cmd/goimports@latest \
    && go install golang.org/x/tools/gopls@latest \
    && go install github.com/go-delve/delve/cmd/dlv@latest \
    && go install honnef.co/go/tools/cmd/staticcheck@latest \
    && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest \
    && go install github.com/spf13/cobra-cli@latest

# Go用のzsh設定追加
RUN echo '# Go environment' >> ~/.zshrc \
    && echo 'export GOPATH=$HOME/go' >> ~/.zshrc \
    && echo 'export GOBIN=$GOPATH/bin' >> ~/.zshrc \
    && echo 'export PATH=$GOBIN:$PATH' >> ~/.zshrc \
    && echo 'export PATH=/usr/local/go/bin:$PATH' >> ~/.zshrc