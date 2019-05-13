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

### Repository Structure

The source code in this repository is used to build two executables: a "master" which serves the web interface and a "slave" which facilitates communication from slave nodes to the master node and provides other helpers. In a single configuration only one instance of the master executable should be running while the slave executable is expected to run on an arbitrary number of slave nodes.

The directory structure is described in detail below.

```
github.com/octacian/extensus
├── config.example.json      # Example configuration file
├── config.json              # Configuration file for master node
├── master                   # Source for executable to be run on master node
│   ├── commands/            # Shell commands
│   ├── core/                # Core APIs to manage loading a variety of required resources
│   ├── models/              # CRUD database APIs with any additional functionality required
│   ├── routes               # HTTP routes
│   │   ├── routes.go        # Mapping of handler functions to routes
│   │   └── ...              # Files contain exported request handler functions
│   └── template/            # Template parsing, rendering, and related helpers
├── migrations/              # Database migrations structured as required by github.com/octacian/migrate
├── public/                  # Public assets served under the `/public/` route
├── shared/                  # Utility APIs and data structures shared by both master and slave source
├── slave/                   # Source for executable to be run on slave nodes
└── templates/               # Templates formatted for use with html/template
    └── base/                # Base templates for inclusion elsewhere
```
