# updev-go-ex-batch-consumer-service
## setup prroject 
1. clone code ``git clone https://github.com/ksupdev/updev-go-ex-batch-consumer-service.git``
2. go mod inti ``go mod init github.com/ksupdev/updev-go-ex-batch-consumer-service``
3. create ``docker-compose.yml`` for start zookeeper and kafka
    - start docker-compose : ``docker-compose up -d``
        ```powershell
        % docker-compose up -d
        --- waiting complete
        Status: Downloaded newer image for 3dsinteractive/kafka:2.0-custom
        Creating updev-go-ex-batch-consumer-service_zookeeper_1 ... done
        Creating updev-go-ex-batch-consumer-service_kafka_1     ... done

        --- check container status ---
        docker ps

        ```
    - clear docker-compose : ``docker-compose down``

