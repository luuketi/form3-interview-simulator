# Form3 Take Home Exercise

## How to build

```
$ make build
```

## How to run

```
$ make run 
```

### How to send a request and terminate the service

```
$ { make run & } && RUNNING_PID=$! && sleep 1 && echo "PAYMENT|1000" | nc localhost 8080 -q 1 && sleep 1 && kill ${RUNNING_PID}
```

## How to test

```
$ make test
```

## Instructions

Located in `INSTRUCTIONS.md`

## Decisions

Assumptions made about the product are located in `DECISIONS.md`
