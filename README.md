# PicServe

A simple image serving application that randomly picks a photo from a directory.


## Overview
PicServe is a lightweight service designed to serve random pictures and supports basic image manipulation.  It is simple version of [Lorem Picusm](https://picsum.photos/) but hosted locally.

My primary use case is to use it with [HomeAssistant](https://www.home-assistant.io/) and [WallPanel](https://github.com/j-a-n/lovelace-wallpanel).

Note that this I'm not a Go developer and chose Go for practice

## Features

## Building
```bash
# Clone the repository
git clone https://github.com/derhally/picserve.git

# Navigate to project directory
cd picserve

# Install dependencies
mage build
```

## Usage
```bash
# Start the server
picserve -photoDir /mnt/media/photos
```
