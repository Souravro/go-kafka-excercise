#!/usr/bin/env bash
function createContainers() {
  container_dir="containers"
  # run zookeeper
  docker-compose -f $container_dir/zookeeper.yml up -d

  # run kafka broker
  docker-compose -f $container_dir/broker.yml up -d

  # run kafka drop
  docker-compose -f $container_dir/kafkadrop.yml up -d
}

function removeContainers() {
  # remove zookeeper
  docker rm -f zookeeper

  # remove kafka broker
  docker rm -f broker

  #remove kafkadrop
  docker rm -f kafkadrop
}

function createTopic() {
   topic_name="idle_user_details"
   partition_count=1
   replication_factor=1

   # exec into a broker and create the topic.
   docker exec broker \
   kafka-topics --bootstrap-server 0.0.0.0:9092 \
               --create \
               --topic $topic_name \
               --partitions $partition_count \
               --replication-factor $replication_factor
}

# main
if [ $1 == "create" ]; then
   createContainers
elif [ $1 == "remove" ]; then
   removeContainers
elif [ $1 == "create-topic" ]; then
   createTopic
else
   echo "Invalid argument. Expected values are 'create' or 'remove'."
fi