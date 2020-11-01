%% Architecture sequence diagram is written in Mermaid markdown syntax
%% using Mermaid JS https://github.com/mermaid-js/mermaid
sequenceDiagram
    participant T as Twitter
    participant F as FrontEnd
    participant B as BackEnd
    Participant D as Database
    Participant M as MiddleWare
    Participant BT as BrainTree
    T -->> T: User tweets "Hey @charityyeti"
    T ->> B: Twitter sends registered listener event
    B -->> B: Drops event on tweetQueue channel
    B -->> B: Processes Twitter data on event
    B ->> D: Saves new entry to DB
    B -->> T: Replies to invoker with custom donation URL
    T ->> F: User opens custom link and fills out donation fields
    F ->> BT: Front end initiates transaction
    BT -->> F: Brain Tree sends transaction nonce back on request
    F ->> B: Front end sends nonce and required data to backend
    B ->> M: Backend forwards nonce and required data to middleware
    M ->> BT: Middleware forwards nonce to Brain Tree to finalize transaction
    BT -->> M: Brain Tree responds good or bad
    M ->> B: Middleware forwards to backend
    B ->> D: Update Tweet DB collection
    B ->> D: Update Transaction DB collection
    B -->> F: Back end sends status to front end
    B -->> T: Back end tweets successful donation to the original tweet