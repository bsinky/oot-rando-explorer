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

# 4.0.0+ versions
for version in `git tag -l`; do
	build_name=`git grep PROJECT_BUILD_NAME $version -- CMakeLists.txt | awk '{ print $2 " " $3 }'`
	if [[ $build_name ]]; then 
		build_name_stripped=`echo $build_name | sed s/\"//g`
		echo "$build_name_stripped ($version)" >> $outfile
	fi
done
