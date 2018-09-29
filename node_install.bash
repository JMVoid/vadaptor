#!/bin/bash
v2_ver=3.45
extract() {
        tar zxvf vadaptor-linux.tar.gz --remove-file
        unzip v2ray-linux-64.zip
        mv v2ray-${v2_ver}-linux-64 v2ray-core
        mv *.json ./v2ray-core/
}

downrun(){

va_ver=`curl --silent "https://api.github.com/repos/JMVoid/vadaptor/releases/latest" |  grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'`
wget https://github.com/JMVoid/vadaptor/releases/download/$va_ver/vadaptor-linux.tar.gz

if [[ $? -eq 0 ]]; then
        v2_ver=`curl --silent "https://api.github.com/repos/v2ray/v2ray-core/releases/latest" |  grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'`
        wget https://github.com/v2ray/v2ray-core/releases/download/$v2_ver/v2ray-linux-64.zip

        if [[ $? -eq 0 ]]; then
                #extract file
                extract
        else
                echo "fail to get vadaptor, exit"
                exit 1
        fi
else
        echo "fail t get v2ray-core, exit"
        exit 1
fi
}

downrun