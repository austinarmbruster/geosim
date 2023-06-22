

PUT aircraft
{
    "mappings": {
        "properties": {
            "@timestamp": {
            "type": "date_nanos"
            },
            "location": {
            "type": "geo_point"
            },
            "name": {
            "type": "keyword"
            },
            "tag": {
            "type": "keyword"
            }
        }
    }
}