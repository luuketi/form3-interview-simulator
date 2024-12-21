# Decisions

- I avoided handling errors from closing connections and also testing some error cases as the service would only log the error.
In a production service such errors would emit metrics to DataDog or New Relic, and the unit tests would validate a metric is emitted in these scenarios.
