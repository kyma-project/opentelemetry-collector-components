# Singleton Receiver Creator

| Status      |                            |
|-------------|----------------------------|
| stability   | alpha: metrics             |
| Code Owners | kyma-project/observability |

Singleton Receiver Creator is an OTel Collector receiver that instantiates another receiver based on the leader election status. It is useful when you want to have a single instance of a receiver running in a cluster. The receiver that gets the lease is created and executed.
This implementation solves the problem mentioned in the following [issue](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/32994). In the future, this implementation can be used as a solution to the aforementioned issue.

## How It Works

It utilizes leader election to determine which instance should be the leader. The instance that wins the lease becomes the leader and starts the underlying sub-receiver. When the leader loses the lease, it stops the receiver and waits until it acquires the lease again.

## Configuration

Below is an example of the configuration:

```yaml
receivers:
  singleton_receiver_creator:
    auth_type: serviceAccount
    leader_election:
      lease_name: foo
      lease_namespace: bar
      lease_duration: 15s
      renew_deadline: 10s
      retry_period: 2s
    receiver:
      otlp:
        protocols:
          grpc:
            endpoint: ""
```
The configuration consists of two parts:
1. The leader election configuration.
2. The receiver configuration.

### Leader Election Configuration
| configuration       | description                                                                     | default value      |
|---------------------|---------------------------------------------------------------------------------|--------------------|
| **lease_name**      | The name of the lease object.                                                   | singleton-receiver |
| **lease_namespace** | The namespace of the lease object.                                              | default            |
| **lease_duration**  | The duration of the lease.                                                      | 15s                |
| **renew_deadline**  | The deadline for renewing the lease. It should be less than the lease duration. | 10s                |
| **retry_period**    | The period for retrying the leader election.                                    | 2s                 |

`auth-type` can be type `serviceAccount`, `kubeConfig`.

### Receiver Configuration
The **name** field specifies the name of the receiver that needs to be created when the instance becomes the leader, followed by the configuration of the receiver that needs to be created.


### Multiple Receivers with Singleton Receiver Creator
To leverage Singleton Receiver Creator with multiple sub-receivers, you must repeat the leader election configuration for each receiver. This is a conscious design decision where each receiver must create and acquire its own lease resource. The main motivation for the decision is that we can divide the load of the Pod where the receiver is running, as running multiple receivers on the same Pod would increase the load on the Pod.


## How to Test

1. Run the following command to deploy the application:

```bash
kubectl apply -f deploy/kube/rbac.yaml
kubectl apply -f deploy/kube/collectors-with-leader.yaml
```

2. Run the following command to check the status of the deployment:

```bash
stern collector -n default
```
