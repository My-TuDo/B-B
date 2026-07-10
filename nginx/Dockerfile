FROM nginx:alpine

ARG APK_MIRROR=mirrors.aliyun.com
RUN if [ -n "${APK_MIRROR}" ]; then \
      sed -i "s/dl-cdn.alpinelinux.org/${APK_MIRROR}/g" /etc/apk/repositories; \
    fi

RUN apk add --no-cache openssl

# Generate self-signed certificate on startup
COPY ssl/generate.sh /docker-entrypoint.d/generate-ssl.sh
RUN chmod +x /docker-entrypoint.d/generate-ssl.sh

COPY nginx.conf /etc/nginx/nginx.conf

EXPOSE 80 443

CMD ["nginx", "-g", "daemon off;"]
