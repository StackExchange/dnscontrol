D("example.com","none"
  // hugh@example.com -> c93f1e400f26708f98cb19d936620da35eec8f72e57f9eec01c1afd6._openpgpkey.example.com -> testing123 -> dGVzdGluZzEyMw==
  , OPENPGPKEY({local:"hugh@", digest:"dGVzdGluZzEyMw=="})
  // mickey.mouse@example.com -> b0c08814687e9ae61ad8ed32b2ef16a3ddcc4ac3bcd1aa8cb4bac73a._openpgpkey.example.com -> testing123 -> dGVzdGluZzEyMw==
  , OPENPGPKEY({local:"mickey.mouse@", digest:"dGVzdGluZzEyMw=="})
  // who . the . fuck . does . this@example.com -> 44bc213e1ff3ece44998c10be434c470421884e6d44c15abff88e09c._openpgpkey.example.com -> testing123 -> dGVzdGluZzEyMw==
  , OPENPGPKEY({local:"who . the . fuck . does . this@", digest:"dGVzdGluZzEyMw=="})
  // I'm.a."noob"@example.com -> 3a14396225556aa5ebc554edd16d5c39d4674d174fcf3a20507ac165._openpgpkey.example.com -> testing123 -> dGVzdGluZzEyMw==
  , OPENPGPKEY({local:"I'm.a.'noob'@", digest:"dGVzdGluZzEyMw=="})
  // `“”‘’‹›«»‘‚„＂〃ˮײ᳓″״‶˶ʺ“”˝‟evilquotes`“”‘’‹›«»‘‚„＂〃ˮײ᳓″״‶˶ʺ“”˝‟@example.com -> db77955230c4e68dd4cb5544d89d602e12c9295f47f8e6c107ae01ba._openpgpkey.example.com -> testing123 -> dGVzdGluZzEyMw==
  , OPENPGPKEY({local:"`“”‘’‹›«»‘‚„＂〃ˮײ᳓″״‶˶ʺ“”˝‟evilquotes`“”‘’‹›«»‘‚„＂〃ˮײ᳓″״‶˶ʺ“”˝‟@", digest:"dGVzdGluZzEyMw=="})
  // alice@example.com -> 2bd806c97f0e00af1a1fc3328fa763a9269723c8db8fac4f93af71db._openpgpkey.example.com
  , OPENPGPKEY({local:"alice@", digest:"\
  -----BEGIN PGP PUBLIC KEY BLOCK-----\
  \
  mDMEZCMu8xYJKwYBBAHaRw8BAQdAH4FTbN/H5SoMBl9Ez2cFQ1NuzymK894fq2ff\
  sYDvRkG0EWFsaWNlQGV4YW1wbGUuY29tiJYEExYKAD4CGwMFCwkIBwMFFQoJCAsF\
  FgIDAQACHgECF4AWIQRjw8oAQytQxDz5Q/Io7xpohfeBngUCZCMv5gUJAAk7ZgAK\
  CRAo7xpohfeBnlmVAP9k0slIpLwddCD1bZ9qVjqzNcS743OIDny7XuH6x02L2wEA\
  wxqAotO7/oUm0L4wyYR6hvGlhuGMSZXc9xMwZ1wVcA8=\
  =vHSO\
  -----END PGP PUBLIC KEY BLOCK-----\
  "})
  // 麻衣子@example.com -> 2bb5bc4202aaecd48dcb54967c8e7f1b7574a436f04e0d15534b20e5._openpgpkey.example.com first target
  , OPENPGPKEY({local:"麻衣子@", digest:"\
  mDMEZCMxgRYJKwYBBAHaRw8BAQdA/fgtlQjGflt2MUMWhRZRnH5Hg+BY9sQTeePm\
  qqUs+lK0Fem6u+iho+WtkEBleGFtcGxlLmNvbYiWBBMWCgA+AhsDBQsJCAcDBRUK\
  CQgLBRYCAwEAAh4BAheAFiEEIWsEkWx5wygGCb61+tJ3q3m88E0FAmQjMbMFCQAJ\
  OqwACgkQ+tJ3q3m88E0z4gEAtowKJMPefyV5YCW8VubgXK7Fa+hjwXOPSsHnEnJw\
  9pUBAL+VZvNZv/VZvyGGMd31Yivqerzl6q+VIkZ6XffVb2AB\
  =sRIg"})
  // 麻衣子@example.com -> 2bb5bc4202aaecd48dcb54967c8e7f1b7574a436f04e0d15534b20e5._openpgpkey.example.com second target
  , OPENPGPKEY({local:"麻衣子@", digest:"\
  -----BEGIN PGP PUBLIC KEY BLOCK-----\
  \
  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\
  =mCoa\
  -----END PGP PUBLIC KEY BLOCK-----\
  "})
  // Η.μπύρα.είναι.νόστιμη@example.com -> ab697306e8b913723a9606e97228cce041898d478a89fc7f4d07509f._openpgpkey.example.com -> testing123 -> dGVzdGluZzEyMw==
  , OPENPGPKEY({local:"Η.μπύρα.είναι.νόστιμη@", digest:"dGVzdGluZzEyMw=="})
);