%% Architecture sequence diagram is written in Mermaid markdown syntax
%% using Mermaid JS https://github.com/mermaid-js/mermaid
sequenceDiagram
    participant T as Twitter
    participant F as FrontEnd
    participant B as BackEnd
    Participant D as Database
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
    F ->> B: Front end sends nonce to backend
    B ->> BT: Back end forwards nonce to Brain Tree to finalize transaction
    BT -->> B: Brain Tree responds good or bad
    B ->> D: Update DB entry with donation value
    B -->> F: Back end sends status to front end
    B -->> T: Back end tweets successful donation to the original tweet