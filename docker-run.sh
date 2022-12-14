#!/usr/bin/env bash
function createContainers() {
  # run zookeeper
  docker run -d \
             -e ZOOKEEPER_CLIENT_PORT=2181 \
             -e ZOOKEEPER_TICK_TIME=2000 \
             --name zookeeper \
             --hostname zookeeper confluentinc/cp-zookeeper:7.3.0

  # run kafka broker
  docker run -d \
             -e KAFKA_ZOOKEEPER_CONNECT=zookeeper:2181 \
             -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092,PLAINTEXT_INTERNAL://broker:29092 \
             -e KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=PLAINTEXT:PLAINTEXT,PLAINTEXT_INTERNAL:PLAINTEXT \
             -e KAFKA_BROKER_ID=1 \
             -e KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1 \
             -e KAFKA_TRANSACTION_STATE_LOG_MIN_ISR=1 \
             -e KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR=1 \
             -p 9092:9092 \
             --name broker \
             --hostname broker confluentinc/cp-kafka:7.3.0

  # run kafka drop
  docker run -d --rm -p 9000:9000 \
        -e KAFKA_BROKERCONNECT=broker:29092 \
        -e JVM_OPTS="-Xms32M -Xmx64M" \
        -e SERVER_SERVLET_CONTEXTPATH="/" \
        --name kafkadrop \
        obsidiandynamics/kafdrop:latest
}

function removeContainers() {
  # remove zookeeper
  docker rm -f zookeeper

  # remove kafka broker
  docker rm -f broker

  # remove kafkadrop
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
