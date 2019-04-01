<h1 align="center">Project Extensus</h1>
<p align="center">An extendable and lightweight systems administration interface.</p>

**Note:** This README does not yet reflect upon the full intentions of Project Extensus and is subject to change.

### Development

#### Key Stages

When developing any major feature, a general order will be followed:
1. Implement Core APIs for the feature
2. *Preferably* implement unit tests for the new Core APIs
3. *Optionally* add a command to the command line interface
4. Implement GraphQL API endpoints connecting to the new Core APIs
5. *Preferably* implement unit tests for the new GraphQL API endpoints
6. Implement front-end views that utilize the GraphQL and Core APIs

#### Short-Term Roadmap

Below is a short-term list of the most prominent tasks going forward. It considers the key stages described above, but does not entirely follow them given that the workflows required for certain stages do not yet exist.
1. Command-Line Interface
2. Core APIs and commands to manage users and permissions
3. GraphQL APIs for user management
4. Login and user management via VueJS front-end
5. Communication between master node (i.e. the web interface) and slave nodes
