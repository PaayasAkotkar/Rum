## Contributing to  RUM 😄

Hey there!!! Welcome!! Thanks for considering to contribute to the RUM. Rum is created on the ideology to manage many models and keep the structure simple than complex. I built idea upon how I viewed the RAG system and how I can make it resuable and control the client's call through registered func.



## Status 💡

Rum/Server:
- **Budget.go:**
                  - implements the budget management.
- **Cluster.go:**
                  - work-in-progress 🚧
- **Dispatch.go**
                  - core to handle the function registry.
- **Handler.go:** 
                  - implements the manipulation of the stream
- **Hub.go:**
                  - concurrently running fetching the channels and dispatching them to the handlers.
- **Kit.go:**
                  - core to handle the descibe the profile
- **Light.go:**
                  - internal pub-sub system.
- **main.go:**
                  - implements the server that handles the gRPC calls -> keeps the hub awake.
- **metric.go:**
                  - implements the log data.
- **profile.go:**
                  - core of the idea that stores the created profiles and modficaiton.
- **register:**
                  - generic function that handle the services to be trigger and pass the result as per the func implemetaiton.
- **rum.go:**
                  - base of the idea that uses all the structs and implements the logic.
- **rumio.go:**
                  - provides writer that writes the result and triggers the events.
- **sequence.go:**:
                  - a simple struct to handle the profile management.
- **service.go:**
                  - implements the description of the profile events.
- **slate.go:**
                  - still trying to figure 😅.
- **store.go:**
                  - implements the search-engine and profile storage.
- **time-format.go:**
                  - implements the description of the events & profile management.
Rum/client:
- **client.go:**
                  - implements the client that calls the server.

Search-Manager:
- **search-manager.go:**
                  - core implementation of the milvus search engine.

- **engine.go:**
                  - main entry point of the search engine.



<div>
<b>Quick start:</b>
<br>
1.Config the minIO user & pass in the compose.yml.
<br>
2.Run the play.bat
<br>
3.you'll see how the registered func actually works
</div>