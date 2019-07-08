#!/usr/bin/env bash

header1="Content-type: application/json"
url="http://0.0.0.0:5984/chain/invoke"

postdata='
{
  "Args":["key1","rec6"]
}
'
#echo $postdata
postdata=`echo $postdata | sed "s/ //g"`
#postdata=`echo $postdata | sed "s/" "/''/g"`
echo $postdata

curl -l -H "$header" -X POST -d $postdata $url
