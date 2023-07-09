# Image Classification

The benchmark implements `Resnet50 model inference` on `Imagenet2012 val` dataset to do image classification. The implementation references [MLPerf Edge Inference Suite](https://github.com/mlcommons/inference/tree/master/vision/classification_and_detection).

The benchmark is implementated in Python.

## Running the benchmark locally(using docker)
1. Start function using docker compose
```bash
docker-compose -f ./yamls/docker-compose/dc-classification.yaml up
```
2. In a new terminal, invoke the interface function with grpcurl
```bash
./tools/bin/grpcurl -plaintext localhost:50000 helloworld.Greeter.SayHello
```

The output includes actual QPS, mean and total inference time, number of queries and tiles.
Hers's an example of benchmark output:
```
TestScenario.SingleStream qps=37.27, mean=0.0263, time=2.683, queries=100, tiles=50.0:0.0263,80.0:0.0264,90.0:0.0265,95.0:0.0268,99.0:0.0271,99.9:0.0290
```

### Parameters
Here are some parameters can be modified in `yamls/docker-compose/dc-classification.yaml` file:
- `--count`: limits the number of items in the dataset used for accuracy pass
- `--time`: limits the time the benchmark runs
- `--scenario`: {'SingStream','Offline'}. Offline means send all queries at one time, SingleStream means send the next query as soon as previous one is completed
- `--threads`: number of worker threads to use (default: the number of processors in the system)
- `--qps`: expected QPS
- `--max-latency`: 
comma separated list of which latencies (in seconds) we try to reach in the 99 percentile (deault: 0.01,0.05,0.100).
- `--max-batchsize`: 
maximum batchsize we generate to backend (default: 128)
