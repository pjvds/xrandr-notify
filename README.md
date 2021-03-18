# xrandr-notify

Subscribe to monitor resolution and other display property changes.

This can be usefull if you want to execute scripts or commands whenever you resolution changes.

## Reload background on resolution change

``` bash
xrandr-notify | while read -r events; do
  feh --bg-fill $current
done
```
