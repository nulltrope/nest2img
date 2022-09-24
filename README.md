# nest2img

`nest2img` is a simple CLI utility for grabbing an image from Nest live video link and saving it to a local `.png` or `.jpeg` file. 

## Compatibility

`nest2img` is compatible with Nest cameras which are managed with the Nest app and can share videos with a public or password-protected link such as `https://video.nest.com/live/<token>`.  See this [Google Nest support thread](https://support.google.com/googlenest/answer/9227530?hl=en) for a comparison of camera sharing options.

Unfortunately it is not compatible with cameras shared via the Google Home app, since those use different authentication and API's, however contributions are welcome!

## Installation

For now, the recommended installation method is using `go install`:
1. Ensure you have a relatively-recent version of Go installed locally (1.18 or higher recommended)
2. Ensure you have `$GOBIN` somewhere on your `$PATH` 
3. Run `go install github.com/nulltrope/nest2img` to install `nest2img` to your `$GOBIN`

## Usage

At a minimum, you'll need your camera's `token` and `password` (if the live link is password-protected). 

To find the token, copy the last portion of your live link which will look something like `https://video.nest.com/live/<token>`. 

Now, run `nest2img` with the token and password (if required):
```
nest2img -token <token> -password <password>
```

If everything worked, you should see an image `out.png` in the current directory with a snapshot from your camera's live feed. Congrats!

### Full CLI Flags

Here's a full output of all CLI flags available to `nest2img`:
```
  -debug
        enable debug logging
  -out string
        the output file, must end in .png or .jpeg (default "out.png")
  -password string
        the camera's password, if link is password-protected
  -quiet
        disable all logging
  -token string
        the camera's token
  -width int
        the image width in pixels (default 512)
```

## Contributing

This is very much a side project for me, but contributions or feature requests via GitHub issues are very welcome, with the caveat that I may or may not get to all of them :) 

## Future Improvements

Some things I'm considering for the future in no particular order:
1. Home Assistant integration, to be able to pull an image on some trigger condition
2. More intuitive logging and configuration options
3. More modular code structure, split into packages for import by other projects