# NLP and Text Search testing

## Elastic Analyzer approach 

### Exact matching

Analyzer: "simple"

Field definition:
```
"<field>":{
"type":"text",
"analyzer": "simple"
...
```

For exact matching, we use the "harmless" simple analyzer to allow for only some error. We can boost this kind of query a lot, so that exact matches come out on top. Using no analyzer at all is also a possibility here.

### Strict language stemming

Analyzer:
```
"custom_english" : {
  "tokenizer" : "standard",
  "filter" : [
    "english_possessive_stemmer",
    "lowercase",
    "english_stop",
    "english_stemmer"]
}
```

To allow for some more error, we can now use the <field>.english multi-field with a language analyzer.

Field definition:
```
"<field>":{
...
  "english": {
    "type":     "text",
    "analyzer": "custom_english"
...
```

For example, now, for phrase matches with distance, we can use the match_phrase or query_string type of the matchm query and configure a *slop* that defines the maximum allowed distance for a match to show up in the results. Documents with "closer" words should get higher scores. We would boost this query less than the exact matches.

### Synonyms

```
"<field>":{
...
  "synonyms": {
    "type": "text",
    "analyzer" : "synonyms"
  }
...
```

The synonym token filter allows to easily handle synonyms during the analysis process. Synonyms are configured using a configuration file, here the Wordnet Prolog synonym db file.

But lets be careful especially when combining synonyms with query_string query!
https://www.elastic.co/guide/en/elasticsearch/guide/current/multi-word-synonyms.html

Also, Elastic v6 seems to break the shingles/synonym approach allowing us to match multiword synonyms as described here: http://opensourceconnections.com/blog/2016/12/02/solr-elasticsearch-synonyms-better-patterns-keyphrases/ As shingles create a graph, getting the synonyms fails as it only allows a flat terms structure.

### Shingles

```
"<field>":{
...
  "shingles": {
    "type": "text",
    "analyzer" : "shingles"
  }
...
```

Good to increase scores for shingles matching. Need to evaluate usefulness when compared to phrase matching. Also would be good to find a way to combine shingles with synonyms in Elastic v6.

### Keep words

For extractions we might only need to match a specific set of words. Or for synonyms, we only want to keep the synonyms not the full list of originals and synonyms.
http://opensourceconnections.com/blog/2016/12/02/solr-elasticsearch-synonyms-better-patterns-keyphrases/
But keep_words from path does not work, bug??

## OpenNLP Integration

https://github.com/spinscale/elasticsearch-ingest-opennlp

PUT /my-index/my-type/1?pipeline=opennlp-pipeline
{
  "<field>" : "Kobe Bryant was one of the best basketball players of all times. Not even Michael Jordan has ever scored 81 points in one game. Munich is really an awesome city, but New York is as well. Yesterday has been the hottest day of the year."
}

Results in:

```
GET /my-index/my-type/1
{
  "<field>" : "Kobe Bryant was one of the best basketball players of all times. Not even Michael Jordan has ever scored 81 points in one game. Munich is really an awesome city, but New York is as well. Yesterday has been the hottest day of the year.",
  "entities" : {
    "locations" : [ "Munich", "New York" ],
    "dates" : [ "Yesterday" ],
    "names" : [ "Kobe Bryant", "Michael Jordan" ]
  }
}
```

We can add custom models. Be careful, 'kobe bryant' or 'KOBE BRYANT' are not real English words and would not be matched using that standard models.

Also, this runs during ingestion, and OpenNLP is not a spell checker. This means any incorrectly spelled entities will be ignored.


## Sample queries

#### Sample complex query

```
GET <INDEX>/_search?pretty
{
"query": {
"bool": {
"should": [
{
"query_string": {
"query": "(house OR noise OR wrongdoer) AND \"wrongdoer SEMI\"~2",
"fields": [
"<field>^8",
"<field>.english^3",
"<field>.synonyms^2",
"<field>.shingles^2",
"<field>.words^10"
],
"type": "most_fields"
}
},
{
"match_phrase": {
"<field>.shingles": {
"query": "described offender",
"slop": 2,
"boost": 6
}
}
}
],
"minimum_should_match": 1,
"boost": 1
}
}
}
```

####  Sample autocomplete query

```
GET /elastic-test/doc/_search
{
    "query": {
        "match_phrase": {
            "<field>.autocomplete": {
                "query":    "from "
            }
        }
    }
}
```

## TODO
Writing tests for checking user requirements on search result expectations.