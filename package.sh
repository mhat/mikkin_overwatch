TEMP=`pwd`/tmpgopath
LICENSE=MIT

mkdir -p $TEMP/bin
mkdir -p $TEMP/src
mkdir -p $TEMP/pkg

GOBIN=$TEMP/bin GOPATH=$TEMP go get github.com/mhat/mikkin_overwatch

VERSION=1.0

#`$TEMP/bin/uniqush-push --version | sed 's/uniqush-push //'`

BUILD=`pwd`/mikkin-overwatch-$VERSION
mkdir -p $BUILD/opt/mikkin_overwatch/bin
mkdir -p $BUILD/opt/mikkin_overwatch/assets
mkdir -p $BUILD/opt/mikkin_overwatch/templates
mkdir -p $BUILD/etc/mikkin_overwatch/
mkdir -p $BUILD/etc/mikkin_overwatch/dropwizard.d/

ARCH=`uname -m`

cp $TEMP/src/github.com/mhat/mikkin_overwatch/LICENSE $LICENSE
cp $TEMP/bin/mikkin_overwatch $BUILD/opt/mikkin_overwatch/bin
cp $TEMP/src/github.com/mhat/mikkin_overwatch/config/yammer-development-environment.json $BUILD/etc/mikkin_overwatch/server.json
cp -R $TEMP/src/github.com/mhat/mikkin_overwatch/assets $BUILD/opt/mikkin_overwatch
cp -R $TEMP/src/github.com/mhat/mikkin_overwatch/templates $BUILD/opt/mikkin_overwatch

#fpm --verbose -s dir -t rpm -v $VERSION -n mikkin_overwatch --license=$LICENSE --maintainer="Matt Knopp" --vendor "Yammer" --description "Yammer Development Environment Log Tailer" -a $ARCH -C $BUILD .
fpm --verbose -s dir -t deb -v $VERSION -n mikkin_overwatch --license=$LICENSE --maintainer="Matt Knopp" --vendor "Yammer" --description "Yammer Development Environment Log Tailer" -a $ARCH -C $BUILD .

rm -rf $TEMP
rm -rf $BUILD
rm $LICENSE
