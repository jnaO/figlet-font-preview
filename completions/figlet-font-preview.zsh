#compdef figlet-font-preview

_arguments -S \
  '--center[centre the rendered text in the preview]' \
  '--right[right-align the rendered text in the preview]' \
  '*--font-dir=[extra directory to search for fonts]:directory:_files -/' \
  '--version[print version and exit]' \
  '(- *)--help[print usage and exit]'
