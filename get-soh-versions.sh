#! /usr/bin/env bash
# Pass the path to the Shipwright repo as the first argument

touch ./randoseed/versions.txt
outfile=`realpath ./randoseed/versions.txt`
# truncate current file
: > $outfile

if [[ $1 ]]; then
	cd "$1"
fi

echo "Generating $outfile..."

# Pre 4.0.0, build version was coded differently
for version in `git tag -l`; do
	# 3.0.0 was the first release with Randomizer
	# https://www.shipofharkinian.com/changelog#rachael-alfa-3-0-0
	if [[ "$version" == "1.0.0" ]] || [[ "$version" == "2.0.0" ]]; then
		continue
	fi
	build_name=`git grep gBuildVersion $version -- soh/src/boot/build.c | grep -Poh "\".+\""`
	if [[ $build_name ]]; then 
		build_name_stripped=`echo $build_name | sed s/\"//g | sed s/\;//g`
		echo "$build_name_stripped" >> $outfile
	fi
done

NATO=(Alfa Bravo Charlie Delta Echo Foxtrot Golf Hotel India Juliett Kilo Lima Mike November Oscar Papa Quebec Romeo Sierra Tango Uniform Victor Whiskey Xray Yankee Zulu)

# 4.0.0+ versions
for version in $(git tag -l); do
    # Extract the raw line from CMakeLists.txt
    build_name=$(git grep PROJECT_BUILD_NAME "$version" -- CMakeLists.txt | awk -F' ' '{print $2 " " $3}')

    if [[ -n "$build_name" ]]; then
        # Clean up quotes
        build_name_stripped=$(echo "$build_name" | sed 's/\"//g')

	# Shipwright CMakeLists changed starting with 9.0.0,
	# now it uses an extra variable to calculate the
	# patch word.
        if [[ "$build_name_stripped" == *"\${PROJECT_PATCH_WORD}"* ]]; then
            # Extract the patch number (the '2' in '9.0.2')
            # This regex looks for the digit after the second dot
            patch_index=$(echo "$version" | cut -d'.' -f3)

            # Look up the word in our Bash array
            replacement_word=${NATO[$patch_index]}

            # Replace the literal variable string with the actual word
            build_name_stripped=${build_name_stripped//\$\{PROJECT_PATCH_WORD\}/$replacement_word}
        fi

        echo "$build_name_stripped ($version)" >> "$outfile"
    fi
done
