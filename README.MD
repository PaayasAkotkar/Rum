<div align="center">
<img src="./asset/rum.png" width=320 height=320>
</div>
<br>

---

## Theme 🎨
Rum or run your models an architecture where a developer can push profiles that include kits that include the respective services to be run, not one, not two, but as many as they can and monitor those specific profiles by creating a tag mark. 
Profile -> Kits -> Services ->[1…N]

---

## Structure 📝

The idea is kept simple — monitor the attributes and allow a robust hierarchy to manage the models or agents. To make things simple gRPC was introduced so that calls can be made from anywhere to support one single server. With this setup the headache of creating many servers to support different agents or models has been solved and you achieve a simple reusable handler for those models.


The hierarchy is:
Rum[In, Out]
└── Profile
└── Kit
└── Service
└── Dispatcher
└── IRegister[In, Out]  ← your model func lives here

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
4. Paper or Poll blocks until all registered funcs complete
<br>
5. Result carries the full metric report — client decides what to do with it
</p>

<br>

---

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
POST() → write() — runs all registered funcs
<br>
↓
<br>
Publish() — internal pubsub (nanoseconds)
<br>
↓
<br>
Poll() or Paper() — returns IDispatchResult
<br>
↓
<br>
Server owns the result
<br>
</div>
<br>

--- 


## Quick Start


- use example pacakge to run any of the example in the main.go
- run the compile.bat
note: examples are written by the gen-ai 
in-case of any error do report the issue.

---

## Key Design Decisions 💡

| Decision | Reason |
|---|---|
| generic [In, Out] | one family, many models, no type confusion |
| gRPC boundary | client triggers from anywhere, language agnostic |
| client owns metrics | engine runs, client decides what to save |

---

## Insiders 😅
| Feature | Status | Insight |
|---|---|---|
| Cluster | Upcoming | One Cluster to scale the service |
|ServerRUM|Upcoming| Reduce complexity for server->client communication |

---

## New 😎
| Feature | Version | Desc |
|---|---|---|
| RUM | v2 | client struct to build easy, poll to fetch the results without cancel|
| Dog | v2 | now able to target cpu, gpu usage , client struct build easy   |

<br>

---

> [!TIP]
> ### 🤝 Let's Collaborate!
> We have big plans for expanding this ecosystem. If you want to join the core group, contribute features, or chat about backend insights, let’s connect:
> 
> * 💬 **Discord:** `Paayas10`
> * 📧 **Email:** [paayasakotkar@gmail.com](mailto:paayasakotkar@gmail.com)

<br>

--- 

<div align="center">
<footer>Apache License 2026; created with imagination and coded with passion</footer>
</div>

