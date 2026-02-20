use criterion::{black_box, criterion_group, criterion_main, Criterion};
use quantun_crypto::mlkem::MlKemKeyPair;
use quantun_crypto::mldsa::MlDsaKeyPair;
use quantun_crypto::slhdsa::SlhDsaKeyPair;
use quantun_types::{MlDsaVariant, MlKemVariant, SlhDsaVariant};

fn bench_mlkem(c: &mut Criterion) {
    let mut group = c.benchmark_group("ML-KEM");

    group.bench_function("keygen-768", |b| {
        b.iter(|| MlKemKeyPair::generate(black_box(MlKemVariant::MlKem768)).unwrap())
    });

    let kp = MlKemKeyPair::generate(MlKemVariant::MlKem768).unwrap();
    group.bench_function("encapsulate-768", |b| {
        b.iter(|| kp.encapsulate().unwrap())
    });

    let enc = kp.encapsulate().unwrap();
    group.bench_function("decapsulate-768", |b| {
        b.iter(|| kp.decapsulate(black_box(&enc.ciphertext)).unwrap())
    });

    group.finish();
}

fn bench_mldsa(c: &mut Criterion) {
    let mut group = c.benchmark_group("ML-DSA");

    group.bench_function("keygen-65", |b| {
        b.iter(|| MlDsaKeyPair::generate(black_box(MlDsaVariant::MlDsa65)).unwrap())
    });

    let kp = MlDsaKeyPair::generate(MlDsaVariant::MlDsa65).unwrap();
    let msg = b"benchmark message for signing";

    group.bench_function("sign-65", |b| {
        b.iter(|| kp.sign(black_box(msg)).unwrap())
    });

    let sig = kp.sign(msg).unwrap();
    group.bench_function("verify-65", |b| {
        b.iter(|| kp.verify(black_box(msg), black_box(&sig)).unwrap())
    });

    group.finish();
}

fn bench_slhdsa(c: &mut Criterion) {
    let mut group = c.benchmark_group("SLH-DSA");

    group.bench_function("keygen-128s", |b| {
        b.iter(|| SlhDsaKeyPair::generate(black_box(SlhDsaVariant::Sha2_128s)).unwrap())
    });

    let kp = SlhDsaKeyPair::generate(SlhDsaVariant::Sha2_128s).unwrap();
    let msg = b"benchmark message";

    group.bench_function("sign-128s", |b| {
        b.iter(|| kp.sign(black_box(msg)).unwrap())
    });

    group.finish();
}

criterion_group!(benches, bench_mlkem, bench_mldsa, bench_slhdsa);
criterion_main!(benches);
