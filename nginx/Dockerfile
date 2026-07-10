FROM nginx:alpine

RUN apk add --no-cache openssl

# Generate self-signed certificate on startup
COPY ssl/generate.sh /docker-entrypoint.d/generate-ssl.sh
RUN chmod +x /docker-entrypoint.d/generate-ssl.sh

COPY nginx.conf /etc/nginx/nginx.conf

EXPOSE 80 443

CMD ["nginx", "-g", "daemon off;"]
