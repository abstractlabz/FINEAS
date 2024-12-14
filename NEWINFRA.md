
# New ETL & LLM Inference Pipeline

## High-level overview of the new ETL pipeline

```
Topic/Segment request JSON or Data Collection JSON -> Data Aggregation Gateway ->
connection to MongoDB instance -> connection to Kafka cluster -> Processors -> RAG ->
Segment
```

---

## Data Aggregator

The data aggregator is the entry point to the data collection system as well as the component that compartmentalizes the flow of data into dynamically defined data stores (i.e. topics and segments).

- **A topic** is the general meaning of what a record of data is about (i.e. financial info, personal health, school notes, etc).
- **A segment** is an attribute which exists under a topic (i.e. company news, health conditions, science notes).

### Setting up topics and segments for data agnostic capture

A request with a config schema sent to the APIGW will dynamically create and delete Kafka topics mapped to MongoDB collections, segments will be stored as a document list under the topic created in MongoDB.

#### Data Aggregation Gateway Topic JSON Schema
```json
{
  "api_key": api_key,
  "topic": topic,
  "topic_action": "create" | "delete"
}
```
The creation or deletion of a topic via this request will map a MongoDB collection to the creation or deletion of a Kafka topic. Deleting a topic with this request will also delete all corresponding segment data.

#### Data Aggregation Gateway Segment JSON Schema
```json
{
  "api_key": api_key,
  "topic": topic,
  "segment": segment,
  "segment_action": "create" | "delete"
}
```

This request will either create or delete a segment of data under a topic.

---

## Processors

### Processor Config JSON Schema
```json
{
  "topic": topic,
  "segment": segment,
  "data_list": "path_to_data_list_json"
}
```

Processors are data capture APIs. They load in a processor config file during initialization, capture data concurrently, and stream raw data plus other metadata to its configured Kafka topic.

- A data collection request call sent to the data aggregator will stream an alert call to processors via a Kafka topic called "alert" specified by the data collection config JSON.
- Processors will collect and produce a stream of data to the RAG system, which will store the inference data in the desired segment defined by the processor config JSON.
- All processors must load a processor config schema during setup as reference to check if a topic exists and if a segment exists under a particular topic.
- All processors will also load in a data list from the processor config to retrieve parameters for data collection.
- All processors will use worker pools to retrieve parameters from the data list in batches and capture or stream data from those batches concurrently.

### Workflow
1. The raw data with the processor config metadata is consumed by the RAG system for inferences and vectorization.
2. The inference is stored in the segment.

A POST request to the Data Aggregation Gateway with the Data Collection JSON should produce a Data Collection JSON (excluding the API key info) and stream it to a general alert Kafka topic:

- The data contains the information for listening processors to kick off their data collection tasks if they match the correct targets by topic and segment defined by a processor’s config file.
- If a process is listening to the alert stream but the alert is for another topic or segment, the data collection task will not start.
- If the process matches the topic and segment, the data capture process begins.

#### Data Collection Config Schema
```json
[
  {
    "api_key": "api_key",
    "topic": "alert",
    "segment_targets": ["x", "y", "z"] // Array of segments to target in the data collection run
  }
]
```

---

## Utilities Directory

The utilities directory should be used to reduce redundancy for:
- Setting up MongoDB connections.
- Setting up Kafka connections.
- Matching processor configs with alert streams.
- CORS Middleware.

---

## Scheduler

Some processors need to collect data more often than others. There needs to be a continuously running program which makes post requests to the aggregator using the data collection JSON on a scheduled basis. We hold a directory of data collection JSONs. We also have a central scheduling JSON file which holds all the paths to the data collection JSONs and their scheduling pattern.

#### Scheduler JSON Schema
```json
[
  {
    "data_collection_file_path": "path",
    "schedule": 1 // call every 1 day(#)
  }
]
```

- An in-memory Redis cache can be used to store a map between a data collection JSON and the last time it was called (all initialized to null on a server startup).
- The scheduler loops through the scheduling JSON and compares the data collection JSON's scheduling pattern with the current time and the last call in the cache.
- If the comparison is overdue or if the last call in the cache is null for that record, then the scheduler will make a POST request to the aggregator using that data collection JSON.

---

## Retrieval Augmented Generation

The retrieval augmented generation services will work to convert raw data into vectors, run large language model inference services for both the chatbot and analysis features, and a Pinecone vector DB query service for the chatbot. These services will all be containerized together, communicate via Kafka streams, and process them concurrently.

### LLM Inferences
- An LLM analysis service will listen for data being sent to all topics, concurrently consume it, send the data to an inference API, then send the inference to an Inference topic for data ingestion.
- An LLM chat service will listen for data being sent to a chatbot topic, concurrently consume it, send the data to an inference API, then send the inference back to the chatbot query service.

### Data Ingestion
- This service will listen for data being sent to all topics, concurrently consume it, chunk the data, and store the vector data in a Pinecone database.

### Chatbot Query
- This service will take POST requests and send the data to a chatbot topic. This data will be consumed and processed by the LLM chat service and sent back to the client.

---

## Deployment Structure

Instead of a modular monolith architecture like the old Fineas backend, the new ETL and Inference pipeline will be a service-oriented architecture. Deployments will package services under 3 containers defined in a Docker Compose or Akash SDL config file:

1. Data aggregation + Scheduler.
2. Processor Servers.
3. Retrieval Augmented Generation Services.

These services will all be independently containerized using Docker Compose and deployed on the same Akash server.

---

## Future Optimizations

### Hybrid Network Scaling
- Kubernetes cluster can be used to deploy a managed network of processor pods across multiple hosts to concurrently extract, process, and stream data in parallel.
- Kubernetes automatic health checks allow failed pods to automatically restart.
- Auto-scaling allows containers within pods to use the optimal amount of server resources on demand.
- Managed networks allow for the benefits of scaling with the security of central management.

### P2P Network Scaling
- Anybody can spin up a Docker node and could connect to a Fineas data capture network.
- Users can get paid in crypto to host this service.
- Potentially infinite horizontal scaling for processors to collect concurrent data in parallel.
- Security concerns with data fraud could be mitigated with some combination of ZK proofs, BFT consensus, or running data through blockchain.
