<center>
	<img src="./logo.png" alt="Aegis Logo" />
</center>

# Aegis

A web frontend for Git.

## Build

Requires Go v1.24+

``` sh
make all
```

## Installation

(for extra details please visit [](./docs/installation.org))

Run the following command:

``` sh
./aegis -init
```

And choose to start the web installer. It will guide you through the configuration process.

After the configuration process, run:

``` sh
./aegis -config {config_file_path}
```

This would set up everything that's required. After the setup process is completed, run the same command again to start the "main" web server.

This process would show you the password for the admin user, which is randomly generated. You can override this password by running:

``` sh
./aegis -config {config_file_path} reset-admin
```

Currently Aegis only supports the following systems for its components; support for other systems are planned:

+ Main database:
  + PostgreSQL
  + SQLite3
+ Receipt system (see [](./docs/receipt.org)):
  + PostgreSQL
  + SQLite3
+ Session store:
  + In-memory
  + SQLite
  + Redis & Redis-like (ValKey, KeyDB)
  + Memcached
+ Mailer:
  + SMTP
  + GMail (through App Password)


