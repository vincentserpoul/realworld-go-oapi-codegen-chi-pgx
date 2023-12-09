# generating config file with header provided by openobserve

Until openobserve has a clear way to directly connect with golang, we need to connect through otel-collector.
In order to send traces and metrics to openobserve, we need to add a header, which is different everytime.
To get this header, go to the (UI)[http://localhost:5080/web/ingestion/recommended/traces?org_identifier=default], and grab the Basic XXX token.

We then need to inject the Authorization header to the config otel-collector-config.yaml, thanks to a sed command

```bash
export TOKEN_VALUE="XXX" && \
sed "s/HEADER_TOKEN_TO_BE_REPLACED/${TOKEN_VALUE}/g" otel-collector-config.tmpl.yaml > otel-collector-config.yaml && \
chmod 666 ./otel-collector-config.yaml
```
