
--------------------
Exact matching
--------------------
"mo_notes":{
"type":"text",
"analyzer": "simple",

For exact matching, we use the "harmless" simple analyzer to allow for only some error. We can boost this
kind of query a lot, so that exact matches come out on top.

--------------------
Strict stemming for less strict matching
--------------------
"mo_notes":{
...
"english": {
"type":     "text",
"analyzer": "custom_english"

To allow for some
more error, we can use the <field>.english multi-field with a language analyzer.

For example, now, for phrase matches with distance, we can use the match_phrase or query_string type of the match
query and configure a *slop* that defines the maximum allowed distance for a
match to show up in the results. Documents with "closer" words should get higher
scores. We would boost this query less than the exact matches.

--------------------
TBD: Handling singular and plurals scoring exact matches higher. Or does this work already?
--------------------
////


--------------------
Synonyms
--------------------
"mo_notes":{
...
"synonyms": {
"type": "text",
"analyzer" : "synonyms"
},

The synonym token filter allows to easily handle synonyms during the analysis process.
Synonyms are configured using a configuration file, for example the Wordnet Prolog
synonym db file.

But lets be careful especially when combining synonyms with query_string query!
https://www.elastic.co/guide/en/elasticsearch/guide/current/multi-word-synonyms.html

Elastic v6 seems to break the shingles/synonym approach allowing us
to match multiword synonyms as described here:
http://opensourceconnections.com/blog/2016/12/02/solr-elasticsearch-synonyms-better-patterns-keyphrases/
As shingles create a graph, getting the synonyms fails as it only allows a flat terms structure.

--------------------
Shingles
--------------------
"mo_notes":{
...
"shingles": {
"type": "text",
"analyzer" : "shingles"
},

Good to increase scores for shingles matching, TBD better later

--------------------
Patterns
--------------------
pattern_capture with full words does not seem to work?

--------------------
Keep Words
--------------------

For extractions we might only need to match a specific set of words. Or for synonyms,
we only want to keep the synonyms not the full list of originals and synonyms.
http://opensourceconnections.com/blog/2016/12/02/solr-elasticsearch-synonyms-better-patterns-keyphrases/
But keep_words from path does not work, bug??

--------------------
OpenNLP integration
--------------------
https://github.com/spinscale/elasticsearch-ingest-opennlp

PUT /my-index/my-type/1?pipeline=opennlp-pipeline
{
  "mo_notes" : "Kobe Bryant was one of the best basketball players of all times. Not even Michael Jordan has ever scored 81 points in one game. Munich is really an awesome city, but New York is as well. Yesterday has been the hottest day of the year."
}

Results in:

GET /my-index/my-type/1
{
  "my_field" : "Kobe Bryant was one of the best basketball players of all times. Not even Michael Jordan has ever scored 81 points in one game. Munich is really an awesome city, but New York is as well. Yesterday has been the hottest day of the year.",
  "entities" : {
    "locations" : [ "Munich", "New York" ],
    "dates" : [ "Yesterday" ],
    "names" : [ "Kobe Bryant", "Michael Jordan" ]
  }
}

We can add custom models...

Problem is that this runs during ingestion, and OpenNLP is not a spell checker.
This means any incorrectly spelled entities will be ignored.

--------------------
Sample queries
--------------------
Sample complex query

GET <INDEX>/_search?pretty
{
"query": {
"bool": {
"should": [
{
"query_string": {
"query": "(house OR noise OR wrongdoer) AND \"wrongdoer SEMI\"~2",
"fields": [
"mo_notes^8",
"mo_notes.english^3",
"mo_notes.synonyms^2",
"mo_notes.shingles^2",
"mo_notes.words^10"
],
"type": "most_fields"
}
},
{
"match_phrase": {
"mo_notes.shingles": {
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

Sample autocomplete query

GET /elastic-test/doc/_search
{
    "query": {
        "match_phrase": {
            "mo_notes.autocomplete": {
                "query":    "from "
            }
        }
    }
}