# ImgServe

A simple image serving application that randomly picks an image from a directory.


## Overview
ImgServe is a lightweight service designed to serve random images with transformations applied.  It is a simple version of [Lorem Picusm](https://picsum.photos/) but hosted locally.

The primary use case was to use it with [HomeAssistant](https://www.home-assistant.io/) and [WallPanel](https://github.com/j-a-n/lovelace-wallpanel).


>[!WARNING]
> This is a very basic not fully optimized web service built for personal use.  It is not meant to 
> * be a high performance image service
> * serve thousands of images
>
> This was built for fun!  Use at your own risk.

## Installation

1. Download latest [release](https://github.com/derhally/imgserve/releases/latest) for your platform.
1. Uncompress the download

## How To Build
```sh
# Clone the repository
$ git clone https://github.com/derhally/imgserve.git

# Navigate to project directory
$ cd imgserve

# if you don't have mage installed
$ go run mage.go

# if you have mage installed
$ mage
```

## Usage

The service will by default listen on port `8080`

This will launch the service service photos from `/mnt/media/photos` and using the dir `/tmp/image_cache` as temporary storage
for transformations.

```sh
# Start the server
$ imgserve -imageDir /mnt/media/photos -cacheDir /tmp/image_cache
```

## API

The following request gets an image that is resized to a certain width and preserves the aspect ratio of the original image

```
http://<address>:<port>/{width}
```


The following gets an image that is resized to fit into a specific width and heigh

```
http://<address>:<port>/{width}/{height}
```


Pass the query parameter `resizemode` to change the resize mode so the image is filled and cropped into a specific width and heigh

```
http://<address>:<port>/{width}/{height}?resizemode=fill
```

To get a grayscale image, pass either `greyscale` or `grayscale` as a query paramter

```
http://<address>:<port>/{width}/{height}?grayscale=1
```

To apply a blur to the image, pass `blur` as a query paramter with a float value

```
http://<address>:<port>/{width}/{height}?blur=1.5
```

## Configuration

### Images

Use the `-imageDir` flag to specify the directory where to load the images from.

Currently images are only loaded from the root directory.  Loading images from subdirectories is not supported.

### Caching

To avoid repeating the same transformation, ImgServe will save the tranformed images to a cache.

Currently the only supported caching mechanism is local storage. 

Use the `-cacheDir` flag to specify a location to save these images.

The service will hash the values of the requested transformation settings and use the has as a filename for future lookups.  The properties used are

* Filename
* Width
* Height
* Blur value
* Grayscale Enabled/Disabled
* Resize Mode

### SSL

If you need HTTPS, you have a couple of options.  

#### Reverse Proxy
Place the service behind a reverse proxy like nginx, caddy, traefik etc.. (This is untested)

#### Using Certificates

If you have certificates then launch the service with the following options:

`-certFile`
Pass the path to the `.cer` file

`-certKeyFile`
Pass the path to the `.key` file

Example

```sh
$ imgserve -imageDir /mnt/media/photos -cacheDir /tmp/image_cache -certFile /home/user/.certs/my.images.net.cer -certKeyFile /home/user/.certs/my.images.net.key
```

### Logging

You can use the `-logLevel` To control the logging levels of the service.  By default the service runs with log level `INFO`

To turn on debug logging use the value `DEBUG`.

## Limitations

1. The service needs to be restarted to serve any new images added to the images directory.
1. There is no database that tracks the images and their properties.  An image will always be loaded into memory to determine the image dimensions.
1. Only JPEG and PNG images are supported
1. The service will only return jpegs back to the requester

## TODO

1. Detect when new images are added to the images directory
1. Restrict the cache by number of images or total cache size
1. Add option to clear cache on exit/start
1. Add observable support to image store to know when items are added, removed, or cleared
1. Support In-memory cache vs disk
1. Support loading images from subdirectories


## Acknowledgements

* [CHIKAMATSU Naohiro](https://github.com/nao1215) for continuing the work on the [imaging](https://github.com/nao1215/imaging) library 

* [fiber](github.com/gofiber/fiber) library.


## License

This is free and unencumbered software released into the public domain. See [LICENSE](./LICENSE) for details.