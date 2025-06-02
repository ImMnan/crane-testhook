
FROM --platform=linux/amd64 alpine:latest

WORKDIR /usr/local/bin
COPY cranetest .

# Set environment variables
ENV WORKING_NAMESPACE=""
ENV ROLE_NAME=""
ENV ROLE_BINDING_NAME=""
ENV SERVICE_ACCOUNT_NAME=""
ENV KUBERNETES_WEB_EXPOSE_TYPE=""
ENV SV_ENABLE=""
ENV KUBERNETES_ISTIO_GATEWAY_NAME=""
ENV DOCKER_REGISTRY=""
ENV KUBERNETES_WEB_EXPOSE_TLS_SECRET_NAME=""
ENV HTTP_PROXY=""
ENV HTTPS_PROXY=""       
ENV NO_PROXY=""          
# Create a non-root user and group
RUN addgroup -g 1337 -S appgroup && adduser -u 1337 -S appuser -G appgroup \
    && chown appuser:appgroup /usr/local/bin/cranetest \
    && chmod +x /usr/local/bin/cranetest

USER appuser

CMD ["/usr/local/bin/cranetest"]
