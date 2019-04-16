#!/bin/bash
set -e

ARCH_UNAME=`uname -m`
if [[ "$ARCH_UNAME" == "x86_64" ]]; then
	ARCH="amd64"
else
	ARCH="386"
fi

EXT="tar.gz"

if [[ "$OSTYPE" == "linux"* ]]; then
	OS="linux"
	UNCOMPRESSED_FILENAME="sd"
elif [[ "$OSTYPE" == "win32" ]] || [[ "$OSTYPE" == "cygwin" ]] || [[ "$OSTYPE" == "msys" ]] ; then
	OS="windows"
	EXT="zip"
	UNCOMPRESSED_FILENAME="sd.exe"
elif [[ "$OSTYPE" == "darwin18" ]]; then
	OS="darwin"
	UNCOMPRESSED_FILENAME="sd"
else
	echo "No binary available for your OS '$OSTYPE'."
	exit
fi

FILENAME=$OS-$ARCH.$EXT

DOWNLOAD_URL="https://github.com/alexanderConstantinescu/go-speed-dial/releases/download/0.2/$FILENAME"

echo "Downloading speed dial from $DOWNLOAD_URL"

if ! curl --fail -o $FILENAME -L $DOWNLOAD_URL; then
    exit
fi

echo ""
echo "extracting $FILENAME to ./${UNCOMPRESSED_FILENAME}"

if [[ "$OS" == "windows" ]]; then
	echo 'y' | unzip $FILENAME 2>&1 > /dev/null
else
	tar -xzf $FILENAME
fi

echo "removing $FILENAME"
rm $FILENAME
chmod +x ./${UNCOMPRESSED_FILENAME}

move_file () {
	echo Moving file to PATH at $1
	sleep 2 
	mv ${UNCOMPRESSED_FILENAME} $1
	sleep 1
}

move_file_with_privilage () {
	IFS=':' read -ra ADDR <<< "$PATH"
	for i in "${ADDR[@]}"; do
		if [[ -d "/usr/local/bin" && $i == "/usr/local/bin" ]]; then
			move_file $i
			break
		fi
	done
}

setup_bash_completion () {
	BASH_COMPLETION_LOCATION=/etc/bash_completion.d/
	if [[ $OS == "darwin" ]]; then
		BASH_COMPLETION_LOCATION=$(brew --prefix)$BASH_COMPLETION_LOCATION
	fi
	if [[ -d "$BASH_COMPLETION_LOCATION" ]]; then
		curl https://raw.githubusercontent.com/alexanderConstantinescu/go-speed-dial/master/sd.bash-completion >> $BASH_COMPLETION_LOCATION/sd
		chmod 644 $BASH_COMPLETION_LOCATION/sd
		echo "Bash completion for sd has been setup. Please start a new shell for the change to take affect"
	else
		echo "Directory: $BASH_COMPLETION_LOCATION does not exist. Cannot setup bash completion"
	fi
}

if { [ $OS == "linux" ] || [ $OS == "darwin" ]; } && [ "$EUID" -eq 0 ]; then
	move_file_with_privilage
	setup_bash_completion
elif [[ $OS == "windows" ]] && net session &> /dev/null; then
	move_file_with_privilage
fi

echo ""
echo "Speed-dial successfully installed to ./sd in your PATH"
echo "Think about renaming the alias in case you think that I gave it an ugly name: 'echo \"alias WHATEVER_YOU_LIKE=sd\" >> ~/.bashrc'"
echo "In case you rename: please restart the bash session for the change to take effect, either by typing 'bash' after this script exists, or by re-logging in."
exit 0
