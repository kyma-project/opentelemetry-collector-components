# Leader Receiver Creator

Leader Receiver Creator is a OTel Collector receiver that instantiates another receiver based on the leader election status. It is useful when one wants to have a single instance of a receiver running in a cluster. The receiver which gets the lease is created and executed.


## Configuration

Below is an example of the configuration:

```yaml
receivers:
  leader_receiver_creator:
    leader_election:
      lease_name: foo
      lease_namespace: bar
      lease_duration: 15s
      renew_deadline: 10s
      retry_period: 2s
    receiver:
      name: "otlp"
      otlp:
        protocols:
          grpc:
            endpoint: ""
```
The configuration consists of two parts:
1. The leader election configuration.
2. The receiver configuration.

### Leader Election Configuration
| configuration   | description                                                                |
|-----------------|----------------------------------------------------------------------------|
| lease_name      | The name of the lease object.                                              |
| lease_namespace | The namespace of the lease object.                                         |
| lease_duration  | The duration of the lease.                                                 |
| renew_deadline  | The deadline for renewing the lease. It should be less than lease duration |
| retry_period    | The period for retrying the leader election.                               |

### Receiver Configuration
The `name` field specifies the name of the receiver that needs to created when the instance becomes the leader, followed by the configuration of the receiver that needs to be created.

## How to test

1. Run the following command to deploy the application:

```bash
kubectl apply -f deploy/kube/rbac.yaml
kubectl apply -f deploy/kube/collectors
```

2. Run the following command to check the status of the deployment:

```bash
stern collector -n default
```
