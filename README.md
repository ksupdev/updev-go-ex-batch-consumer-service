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
4. run ``go run *.go``

## implement project
1. ``context.go`` จะมีการเพิ่มในส่วน ``ReadInputs() []string`` เพื่อใช้ในการอ่านข้อมูล ที่เป็น collections
2. create ``context_consumer_batch.go`` สำหรับจัดการในส่วนของ value ใน context และเป็น Implementation ของ interface ``IContext`` ใน file context.go
3. create ``microservice.go`` ซึ่งจะเป็นส่วนที่จัดการ function การทำงานของ service ที่เป็น Bath
    - implement ``IMicroservice`` เพื่อเป็นการกำหนด function ของการทงานของ service
    - implement ``ServiceHandleFunc`` ใช้สำหรับ handle การทำงานของแต่ละ service โดยจะออกแบบให้สามารถทำงานได้อย่าง dynamin logic
    - implement ``Microservice`` and ``NewMicroservice`` สำหรับเป็น implementation ของ ``IMicroservice``
    - implement func ``newKafkaConsumer`` สำหรับ create connection ของ Kafka consumer และจะต้องทำการ เรียกใช้งาน lib 
    - implement func ``consumeBatch`` ????

        ```powershell
        % go get github.com/confluentinc/confluent-kafka-go/kafka
        go: found github.com/confluentinc/confluent-kafka-go/kafka in github.com/confluentinc/confluent-kafka-go v1.6.1
        ```
4. create ``batch_event.go`` ซึ่งจะทำหน้าในการ keep data ที่ได้รับมาจาก kafka และค่อยทำการ execute ตาม function โดยจะมีการทำงานในรูปแบบ Queue โดยใน file batch_event.go จะมีการ import package ``github.com/phf/go-queue/queue`` เพื่อใช้สำหรับการจัดการ queue ในการ excute ผลลัพธ์ที่ได้จาก Kafka
    - implement ``Batch`` สำหรับการเก็บข้อมูลที่ได้จาก Kafka โดยเราจะเก็บในรูปแบบ Queue หรือก็คือรับ Response ทั้งหมดเข้ามาก่อนแล้วค่อยเอาไป execute ทีเดียว
    - implement ``NewBatch()`` สำหรับใช้ในการ clear ค่าใน Queue ให้เป็นค่าว่างหรือก็คือการ initial ค่าเข้าไป
    - implement ``Add()`` ใช้สำหรับการใส่ค่าล่งไปใน Queue (FIFO)
    - implement ``Read()`` ใชำสำหรับอ่านค่าออกจาก Queue (FIFO)
    - implement ``Reset()`` ใช้สำหรับเคลียค่าใน Queue (FIFO)
    - implement ``BatchEvent`` struct โดย เราจะใช้ในการจัดการกับข้อมูลที่อยู่ใน ``Batch`` ที่เราเก็บข้อมูลไว้นั้นเอง
    - implement ``Start()`` ......?
```powershell
% go get github.com/phf/go-queue/queue
go: downloading github.com/phf/go-queue v0.0.0-20170504031614-9abe38d0371d
go: found github.com/phf/go-queue/queue in github.com/phf/go-queue v0.0.0-20170504031614-9abe38d0371d
```

5. create ``producer.go``
> Bulk Consumer : จะทำทุกอย่างเหมือนกับ Consumer แต่จะทำที่ละหลายๆ recoard และจะต้องทำควบคู่กับ ระบบอื่นๆที่ support การทำงานแบบ Bulk เช่นถ้า Consumer Loop insert to db แต่ Bulk จะเป็นการเอา Recoard ทั้งหมดและใช้ function bulk ข้อมูลทุกๆ Recoard และทำการ save ลง database ทันที แต่ database นั้นต้อง support bulk function ด้วย

6. Docker down
```powershell
docker-compose down
```

