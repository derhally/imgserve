# ImgServe

A simple image serving application that randomly picks an image from a directory.


## Overview
ImgServe is a lightweight service designed to serve random images with transformations applied.  It is simple version of [Lorem Picusm](https://picsum.photos/) but hosted locally.

My primary use case is to use it with [HomeAssistant](https://www.home-assistant.io/) and [WallPanel](https://github.com/j-a-n/lovelace-wallpanel).


## Features

## Building
```bash
# Clone the repository
git clone https://github.com/derhally/imgserve.git

# Navigate to project directory
cd imgserve

# Install dependencies
mage build
```

## Usage
```bash
# Start the server
imgserve -imageDir /mnt/media/photos -cacheDir /tmp/image_cache
```


## TODO

1. Constrain cache by size or count
2. Add observable support to image store to know when items are added, removed, or cleared
3. In-memory cache