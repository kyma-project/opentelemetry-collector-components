# Singleton Receiver Creator

| Status      |                            |
|-------------|----------------------------|
| stability   | alpha: metrics             |
| Code Owners | kyma-project/observability |

Singleton Receiver Creator is a OTel Collector receiver that instantiates another receiver based on the leader election status. It is useful when one wants to have a single instance of a receiver running in a cluster. The receiver which gets the lease is created and executed.
The problem this implementation solves has been mentioned in following [issue](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/32994). In future this implementation can be used as solution to the aforementioned issue.

## How it works

It utilizes leader election to determine which instance should be the leader. The instance which wins the lease, becomes the leader and starts the underlying sub-receiver. When the leader loses the lease, it will stop the receiver and wait until it acquires the lease again.

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


### Multiple receiver with Singleton Receiver Creator
For leveraging singleton receiver creator with multiple sub receiver, one needs to have to repeat the leader election configuration for each receiver. This is a conscious design decision, where each receiver would have to create and acquire its own lease resource. The main motivation for the decision is that we can divide the load of the pod where the receiver is running as running multiple receivers on same pod would increase the load on the pod.


## How to test

1. Run the following command to deploy the application:

```bash
kubectl apply -f deploy/kube/rbac.yaml
kubectl apply -f deploy/kube/collectors-with-leader.yaml
```

2. Run the following command to check the status of the deployment:

```bash
stern collector -n default
```
