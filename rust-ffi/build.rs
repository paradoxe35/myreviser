use std::env;
use std::path::PathBuf;

fn main() {
    let crate_dir = env::var("CARGO_MANIFEST_DIR").unwrap();
    let output_file = PathBuf::from(&crate_dir).join("bindings.h");

    cbindgen::Builder::new()
        .with_crate(crate_dir)
        .with_language(cbindgen::Language::C)
        .with_pragma_once(true)
        .with_include_guard("MYREVISER_FFI_H")
        .with_documentation(true)
        .with_sys_include("stdint.h")
        .with_sys_include("stdbool.h")
        .with_tab_width(4)
        .generate()
        .expect("Unable to generate C bindings with cbindgen")
        .write_to_file(output_file);

    println!("cargo:rerun-if-changed=src/");
}
