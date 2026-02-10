# Useage: pack-oci.sh image_file <dst>
set -e

imgs=$1
dst=$(realpath $2)


if [[ -z "$imgs" || -z "$dst" ]]; then
    echo "Usage: $0 <image_file> <dst>"
    exit 1
fi


tempdir=$(mktemp -d)
cleanup() {
    rm -rf "$tempdir"
}
trap cleanup EXIT INT TERM

for i in $(cat "$imgs"); do
    echo "Coping image: ${i} to oci:${tempdir}:${i}"
    skopeo copy --insecure-policy docker://${i} oci:${tempdir}:${i} || \
    skopeo copy --insecure-policy docker-daemon:${i} oci:${tempdir}:${i}
done

cd "$tempdir" && tar cf "$dst" *
