# Deploy and Host DataLens on Railway

[![Deploy on Railway](https://railway.com/button.svg)](https://railway.com/new/template/datalens?referralCode=sLjDXb&utm_medium=integration&utm_source=button&utm_campaign=datalens)

[DataLens](https://datalens.tech/) is the open-source business intelligence platform from Yandex: connect your databases, model datasets with calculated fields, build charts in a drag-and-drop wizard and put them on interactive dashboards with selectors, tabs and drill-downs. This template runs the full open-source stack from upstream's own images, with a login, a generated admin password and a demo workbook to explore on first open.

## About Hosting DataLens

The stack is six services: DataLens (the web app, the only public one), US (United Storage, the metadata store), Auth, ControlAPI, DataAPI and Postgres.

- **Upstream's images, pinned to release 2.9.0.** Each service is upstream's `ghcr.io/datalens-tech` image at the versions of the v2.9.0 release, with a thin wrapper that adds only key setup. Nothing moves until you bump a tag.
- **Every key unique to your deploy.** Upstream's compose file ships fixed RSA keys and tokens, the same for every install. Here one generated seed on Postgres is the source of every signing key (user sessions, service-to-service tokens), both master tokens and the key that encrypts saved connection passwords. Each service derives the ones it needs at start, so nothing secret is shared in public.
- **Login on, sign-up off.** The `admin` user is created on first boot with a generated password. Self sign-up is disabled: the admin adds users in Service settings, so nobody else can register on your public URL. No email server needed.
- **A demo workbook.** First boot loads upstream's demo: a sample marketplace dataset with a dashboard showing most chart types. It runs on its own non-superuser Postgres role, so charts built on the demo connection can't read DataLens's own databases.
- **Lean by default.** Upstream's workbook export/import needs two more services (meta-manager and a Temporal server); they're left out to keep the stack smaller. Everything else, including the dataset editor, the chart wizard, QL charts and dashboards, works.

## Common Use Cases

- Self-hosted dashboards over your production Postgres, MySQL or ClickHouse
- Replacing a hosted BI subscription with a team-shared instance you control
- Ad-hoc analysis with calculated fields, window functions and level-of-detail expressions
- Embedding the same metrics in one place for sales, product and operations teams

## Dependencies for DataLens Hosting

- Postgres 16 for metadata, users and the demo data (included, private network only)
- The databases you want to analyze (PostgreSQL, ClickHouse, MySQL, Greenplum, MS SQL Server, Oracle, Trino and others)

### Deployment Dependencies

- [DataLens documentation](https://datalens.tech/docs/en/)
- [DataLens on GitHub](https://github.com/datalens-tech/datalens)
- [Template source on GitHub](https://github.com/nomideusz/datalens-railway)

### Implementation Details

**Sign in.** Open the DataLens service's Railway domain and sign in as `admin` with the `DATALENS_ADMIN_PASSWORD` from the DataLens service's Variables tab. The password is applied on first boot only; change it later under Account in DataLens. The first boot takes a few minutes while the six services build and the demo data loads.

**Add users.** Service settings → Users → Add user. Give people the Viewer, Editor or Admin role.

**Connect your data.** Open a workbook, choose Create → Connection and point it at your database. Databases in the same Railway project are reachable on their private domain (for example `postgres.railway.internal`); for anything else use its public host. DataLens queries your database live, so give it a read-only user.

**Never change `DATALENS_KEY_SEED`** after the first boot. It encrypts saved connection passwords; a new seed makes every existing connection unreadable and signs everyone out.

**Custom domain.** Add the domain to the DataLens service, then set `UI_APP_ENDPOINT` on the Auth service to `https://your-domain`. Sign-in cookies are issued for that host only.

**Workbook export/import** is turned off because it needs upstream's meta-manager and Temporal services, which this template leaves out. Upstream's [docker-compose.yaml](https://github.com/datalens-tech/datalens/blob/main/docker-compose.yaml) shows how they are wired if you want to add them.

**Maps.** Map charts need a Yandex Maps API key: set `YANDEX_MAP_ENABLED=1` and `YANDEX_MAP_TOKEN` on the DataLens service.

**Resources.** RAM_PLACEHOLDER

## Why Deploy DataLens on Railway?

Railway is a singular platform to deploy your infrastructure stack. Railway will host your infrastructure so you don't have to deal with configuration, while allowing you to vertically and horizontally scale it.

By deploying DataLens on Railway, you are one step closer to supporting a complete full-stack application with minimal burden. Host your servers, databases, AI agents, and more on Railway.
