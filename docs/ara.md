*Adaptive Resource Allocation for Kubernetes Pods*

At StitchFix we empower our data scientists to deploy their models and applications end to end without needing engineering skills. To facilitate batch processing we use Flotilla, a task execution service. Flotilla can run jobs on top of Kubernetes or AWS ECS.

One of the problems we faced was how much CPU and memory should we assign to the container pods? The workloads are highly variable on their demands. 

If we give too few resources the jobs may run slower and in the pathological case of running out of memory. If we give too much we are wasting resources and starving other jobs that could potentially be scheduled alongside. 

Solution
The first step was to accurately record the utilization of the resources per pod. We looked at a few different monitoring solutions (kube-state-metrics, Prometheus, and metrics-server). We decided to use the metrics-server since it provided a simple API and tracked the state of the pods in memory. 

```
helm install --name=metrics-server --namespace=kube-system --set args={'--metric-resolution=1s'} stable/metrics-server
```
To instrument fetching the pod metrics, we used the metrics ClientSet. While the job is running, Flotilla fetches the metrics every 2-5 seconds.

If the prior recorded value of memory and CPU are lower than what the Metrics Server is outputting the highest of the two are recorded back with job metadata.

Also, an MD5 checksum of the command and its arguments are stored in the database. This becomes a signature of the job and its resources. 

The core [query for ARA](https://github.com/stitchfix/flotilla-os/blob/master/state/pg_queries.go#L53-L66) and the associated [adapter code](https://github.com/stitchfix/flotilla-os/blob/master/execution/adapter/eks_adapter.go#L269-L301)
