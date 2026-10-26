# #region build
# 非 root（UID 101）で 8080 番を待ち受ける nginx の公式イメージ
FROM nginxinc/nginx-unprivileged:1.30.6-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY index.html /usr/share/nginx/html/index.html
EXPOSE 8080
# #endregion build
