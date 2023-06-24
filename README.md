
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
