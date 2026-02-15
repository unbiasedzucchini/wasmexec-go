(module
  (memory (export "memory") 4)

  (func (export "run") (param $input_ptr i32) (param $input_len i32) (result i32)
    (local $out i32)
    (local $i i32)

    (local.set $out (i32.const 0x20000))

    ;; write output_len as u32 LE at $out
    (i32.store (local.get $out) (local.get $input_len))

    ;; copy input bytes to $out + 4
    (local.set $i (i32.const 0))
    (block $break
      (loop $loop
        (br_if $break (i32.ge_u (local.get $i) (local.get $input_len)))
        (i32.store8
          (i32.add (i32.add (local.get $out) (i32.const 4)) (local.get $i))
          (i32.load8_u (i32.add (local.get $input_ptr) (local.get $i)))
        )
        (local.set $i (i32.add (local.get $i) (i32.const 1)))
        (br $loop)
      )
    )

    (local.get $out)
  )
)
