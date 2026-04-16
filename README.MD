## RUM
<div align="center">
<img src="./image/logo.png" alt="RUM Logo" width="132" height="132">
</div>

## Theme 🎨
Rum provides the robust way to manage many agent models at once.
It allows one single server to handle the flow of stream; the 
call can be managed via client.

## Motivation 🔥

<div>

<div align="center"><b>Beginning....</b></div>

Since the day one when I started researching about RAG and reading research papers from [alphaxiv](https://www.alphaxiv.org/) I noticed that each model has their own attribute than I thought of creating a life cycle to monitor these models and have control when to stop and when to activate. I wanted something that can be controllable since first request without using if else statement.
<br>
<div align="center"><b>Why Milvus as Search Engine?</b></div>

The next thing was to figure out the search engine to allow the user to freely use the search engine without operating anything else. After trying pg-vector, Milvus solved the problem cleanly and also came with MinIO (S3-compatible storage) which was a bonus.

<br>
<div align="center"><b>Structure</b></div>

The idea is kept simple — monitor the attributes and allow a robust hierarchy to manage the models or agents. To make things simple gRPC was introduced so that calls can be made from anywhere to support one single server. With this setup the headache of creating many servers to support different agents or models has been solved and you achieve a simple reusable handler for those models.
<p>
The hierarchy is:
<br>
Rum[In, Out]
<br>
└── Profile
<br>
└── Kit
<br>
└── Service
<br>
└── Dispatcher
<br>
└── IRegister[In, Out]  ← your model func lives here
</p>

<br>
<div align="center"><b>Flow</b></div>

<p align="center">
1. Register the profile [In, Out]
<br>
Note: keep [In & Out] same for each model so that you can push many profiles.
<br>
2. Run the server
<br>
3. Call the server via gRPC client from anywhere
<br>
4. Paper blocks until all registered funcs complete
<br>
5. Result carries the full metric report — client decides what to do with it
</p>

<br>

## Architecture 🏗️
<div align="center">
Client [grpc client anywhere]
<br>
↓ 
<br>
Rum Server [grpc server]
<br>
↓
<br>
Hub() — single event loop
<br>
↓
<br>
POST() → write() — runs all the active registered funcs as per the rankings on set description
<br>
↓
<br>
Light.Publish() — internal pubsub (nanoseconds)
<br>
↓
<br>
Papers() — returns IDispatchResult
<br>
↓
<br>
Server owns the result
<br>
</div>

## Key Design Decisions 💡

| Decision | Reason |
|---|---|
| Generic [In, Out] | One family, many models, no type confusion |
| gRPC boundary | Client triggers from anywhere, language agnostic |
| Internal pubsub | No network hop, nanosecond latency |
| Budget per service | Cost control built in from day one |
| Client owns metrics | Engine runs, client decides what to save |

## Insiders 😅
| Feature | Status | Insight |
|---|---|---|
| Embedder | Upcoming | One single embedder for all models |
| Cluster | Upcoming | One Cluster to scale the service |
|ServerRum|Upcoming| Reduce complexity for server->client communication |
<div align="center">
<footer>@copyright 2026; created with imagination and coded with passion ✏️</footer>
</div>