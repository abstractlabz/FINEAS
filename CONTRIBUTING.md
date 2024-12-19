

### `CONTRIBUTING.md`

---

# Contributing to Fineas.AI

Thank you for considering contributing to Fineas.AI! Your contributions help us improve and grow. Here are guidelines for contributing to the project.

## Table of Contents

1. [Overview](#overview)
2. [Architecture Summary](#architecture-summary)
3. [Contribution Guidelines](#contribution-guidelines)
4. [Setting Up the Development Environment](#setting-up-the-development-environment)
5. [Coding Standards](#coding-standards)
6. [Commit Messages](#commit-messages)
7. [Pull Request Process](#pull-request-process)
8. [Code of Conduct](#code-of-conduct)

## Overview

Fineas.AI is a microservices-based system designed to provide comprehensive financial insights and market analysis. This document outlines the project's architecture, contribution guidelines, and standards to help you get started with contributing effectively.

## Architecture Summary

### Current State

Fineas.AI consists of various components, each responsible for specific functionalities, interacting through well-defined APIs and leveraging external services for data and processing.

#### Components and Workflow

1. **Aggregator Service:**
   - Central service handling requests for aggregated financial information.
   - Coordinates data retrieval from various sub-services (Stock, News, Financials, Description, Technical Analysis).
   - Integrates data and responds to client requests.

2. **Sub-Services:**
   - **Stock Service:** Fetches stock prices and calculates percent changes.
   - **News Service:** Scrapes news articles related to specific tickers.
   - **Financials Service:** Retrieves and processes financial statements.
   - **Description Service:** Provides detailed company descriptions.
   - **Technical Analysis Service:** Analyzes stock data technically.

3. **Data Storage:**
   - **MongoDB:** Central data storage for logging and maintaining service data.
   - **Pinecone:** Used for vector storage, enabling efficient data retrieval and embedding.

4. **AI and Embeddings:**
   - **Claude:** Provides AI-driven responses and embeddings for chatbot functionalities.
   - **Chatbot Data Ingestor:** Processes and ingests data for the chatbot.
   - **Chatbot Query Service:** Handles user queries and interacts with Pinecone and OpenAI.

5. **Middleware:**
   - **CORS Middleware:** Handles cross-origin requests.
   - **Retrieval Service:** Fetches aggregated data from MongoDB for client use.

6. **External APIs:**
   - **Polygon API:** Primary data source for financial information.

### Advantages

- **Modular Architecture:** Each service is responsible for a specific functionality, promoting separation of concerns and easier maintenance.
- **Parallel Processing:** Data for stock tickers are computed in parallel within the aggregator using goroutines.
- **Security:** Services are secured using hashed pass keys, ensuring only authorized requests are processed.

### Disadvantages

- **High Coupling:** Services are highly coupled to retrieving and preprocessing data from specific hardcoded data sources.
- **Latency:** Inter-service communication over the network can introduce latency and affect response times.
- **Resource Intensive:** Code redundancy between services leads to higher memory and compute overhead.

### Recommendations

- **Optimize Inter-Service Communication:** Use asynchronous communication methods such as Kafka consumers to pre-process and format API response data.
- **Implement Caching:** Use in-memory caching like Redis to reduce load on database queries and improve response times.
- **Use Container Orchestration:** Implement container orchestration platforms like Kubernetes to manage and scale services efficiently.
- **Optimize Database Usage:** Use indexing and query optimization techniques to improve database performance.
- **Monitor and Analyze Metrics:** Implement comprehensive monitoring and logging to analyze performance metrics and identify areas for improvement.
- **Cost Management:** Regularly review and optimize cloud resource usage to reduce operational costs.

### KPIs to Measure

- **Response Time:** Measure the average response time for each service to identify bottlenecks.
- **Request Throughput:** Track the number of requests handled by each service per second to assess load handling capacity.
- **Error Rates:** Monitor the rate of errors and failures across services to ensure reliability.
- **Resource Utilization:** Measure CPU, memory, and network usage for each service to optimize resource allocation.
- **Latency:** Track the time taken for inter-service communication to identify and reduce latency.
- **Uptime:** Monitor the uptime and availability of each service to ensure high availability.

## Contribution Guidelines

### Coding Standards

- **Follow the existing code style:** Maintain consistency with the project's code style.
- **Write tests:** Ensure that new features and bug fixes are covered by appropriate tests.
- **Document your code:** Provide clear and concise comments and documentation for your code.

### Commit Messages

- **Use meaningful commit messages:** Clearly describe the changes in your commit.
- **Reference issues:** If your commit addresses an issue, reference it in the commit message (e.g., `Fixes #123`).

### Pull Request Process

1. **Fork the repository:**
   Create a fork of the repository on GitHub.

2. **Create a new branch:**
   ```sh
   git checkout -b feature/your-feature-name
   ```

3. **Make your changes:**
   Implement your feature or fix and commit your changes.

4. **Push to your fork:**
   ```sh
   git push origin feature/your-feature-name
   ```

5. **Create a pull request:**
   Go to the repository on GitHub and create a pull request to merge your changes into the main branch.

### Code of Conduct

- **Be respectful:** Treat all contributors with respect and courtesy.
- **Collaborate:** Work together to improve the project, and be open to constructive feedback.

By following these guidelines, you help ensure that Fineas.AI remains a robust and high-quality project. Thank you for your contributions!

---