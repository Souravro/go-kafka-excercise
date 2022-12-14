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

# main
if [ $1 == "create" ]; then
   createContainers
elif [ $1 == "remove" ]; then
   removeContainers
else
   echo "Invalid argument. Expected values are 'create' or 'remove'."
fi