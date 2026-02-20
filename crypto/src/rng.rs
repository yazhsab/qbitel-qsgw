//! RNG adapter bridging `getrandom` to `rand_core 0.10` traits.
//!
//! The PQC crates (ml-kem, ml-dsa, slh-dsa) use `rand_core 0.10`'s
//! `CryptoRng` trait (via `signature 3.x`), while our workspace uses
//! `rand 0.8` / `rand_core 0.6`. This adapter creates a
//! `CryptoRng`-compatible wrapper using `getrandom`.

use core::convert::Infallible;

// Access rand_core 0.10 traits through the signature crate's re-export
use signature::rand_core as rc10;

/// OS-backed cryptographically secure RNG for use with PQC crate APIs.
///
/// Implements `rand_core 0.10` traits using `getrandom::fill()` as the
/// entropy source. This bridges the gap between our `rand 0.8` workspace
/// and the PQC crates that require `rand_core 0.10`'s `CryptoRng`.
///
/// # Panics
///
/// Operations will panic if the OS entropy source is unavailable. This is
/// considered unrecoverable -- a system without a working RNG cannot safely
/// perform any cryptographic operations.
pub struct PqcRng;

impl rc10::TryRng for PqcRng {
    type Error = Infallible;

    fn try_next_u32(&mut self) -> Result<u32, Self::Error> {
        let mut buf = [0u8; 4];
        // SAFETY: getrandom::fill uses the OS CSPRNG. If it fails, the system
        // is in an unrecoverable state and panicking is the correct behaviour
        // for a cryptographic RNG -- continuing with bad entropy would be worse.
        getrandom::fill(&mut buf).expect("OS entropy source unavailable — cannot proceed safely");
        Ok(u32::from_le_bytes(buf))
    }

    fn try_next_u64(&mut self) -> Result<u64, Self::Error> {
        let mut buf = [0u8; 8];
        getrandom::fill(&mut buf).expect("OS entropy source unavailable — cannot proceed safely");
        Ok(u64::from_le_bytes(buf))
    }

    fn try_fill_bytes(&mut self, dest: &mut [u8]) -> Result<(), Self::Error> {
        getrandom::fill(dest).expect("OS entropy source unavailable — cannot proceed safely");
        Ok(())
    }
}

// Marker trait: this RNG is cryptographically secure
impl rc10::TryCryptoRng for PqcRng {}
