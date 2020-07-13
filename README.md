# Prometheus service discovery for Exoscale Instance Pools

This is a service discovery agent for Prometheus that uses Exoscale instance pools.

You can run it using Docker:

```
docker run \
    # Run in background
    -d
    # Mount the data directory
    -v /srv/service-discovery:/var/run/prometheus-sd-exoscale-instance-pools \
    janoszen/prometheus-sd-exoscale-instance-pools \
    # Provide the Exoscale API key here:
    --exoscale-api-key EXO... \
    # And the secret:
    --exoscale-api-secret ... \
    # Run the `exo zone` command to get this value
    --exoscale-zone-id 4da1b188-dcd6-4ff5-b7fd-bde984055548 \
    # Run the `exo instancepool list` command to get this value:
    --instance-pool-id ...
```

