#### cert

```
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes \
    -subj "/C=US/ST=California/L=Any City/O=Poolside/OU=Org/CN=*.poolside.dev"
```
