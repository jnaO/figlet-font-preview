# fish completion for figlet-font-preview
complete -c figlet-font-preview -f
complete -c figlet-font-preview -l center -d 'centre the rendered text in the preview'
complete -c figlet-font-preview -l font-dir -r -x -a '(__fish_complete_directories)' -d 'extra directory to search for fonts'
complete -c figlet-font-preview -l version -d 'print version and exit'
complete -c figlet-font-preview -l help -d 'print usage and exit'
