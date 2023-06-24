# Automate Most of the Tracking Containment Exampmle

Creating a demo of the tracking containment environment is simplified with a little bit of automation. 
This directory has the Terraform logic to create two clusters:

1. A "user" cluster, which is the cluster that we want to run the Tracking Containment rules.
1. An "olly" cluster, which is the cluster in which we will perform the audit & logging work.

## Setup

The Terraform plugins include the `elastic/ec` and `elastic/elasticstack`. 
The `elastic/ec` plugin facilitates creating the clusters. 
The `elastic/elasticstack` plugin provides several helpers to manage the contents of a cluster.

In order to create the clusers, you will need an Elastic Cloud API Key.

1. Create an API key in Elastic Cloud
   - Navigate to the API Key page
   - Click `Create API Key`
   - Name the Key, e.g. `audit-demo`
   - Click `Create API Key`
   - Copy the API key to somewhere safe, since it will not be shown again.
1. You will need to set the API key as the value of the `EC_API_KEY` environment variable, e.g. `export EC_API_KEY=<your api key>`

In addition to the API key, you will need to have Terraform. Hashicorp, the creators of Terraform, have instructions for [installing Terraform](https://developer.hashicorp.com/terraform/downloads).

1. Install `terraform`
1. In the automation directory, run `terraform init`
   - This will get the plugins that are required to run the install

## Installation

With the API key and terraform available, you can create the clusters and start the data flow by simply running `make`

## Sample State Queries

```
GET .kibana_task_manager/_search?size=40

GET .kibana-event-log-*/_search
{
  "query": {
    "bool": {
      "must": [
        {
          "match": {
            "rule.name": "filtered"
          }
        },
        {
          "match": {
            "event.action": "execute"
          }
        },
        {
          "match": {
            "kibana.alert.rule.execution.metrics.alert_counts.active": 1
          }
        }
      ]
    }
  }
}
```
