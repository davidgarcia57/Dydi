import React from 'react';
import Svg, { G, Circle } from 'react-native-svg';

// "Al corriente": el anillo cerrado, los tres arcos en salvia, las bolas en
// reposo. Misma familia que SoloSquadFigure y el logo (r=70 en viewBox 240,
// trazo 20 = 8.3% del diámetro, cap redondo, dasharray "115 325", arcos a
// 0/120/240).
//
// La decisión de diseño que la distingue: aquí NO hay elemento fuera de registro.
// En el resto del sistema el arco desplazado significa "algo se resbaló"; un
// anillo perfectamente alineado es justo lo contrario, y es lo que significa no
// tener deudas. La ausencia del desajuste es el mensaje.
//
// Salvia y no terracotta porque aquí el color sí es semántico: sage = al
// corriente, el mismo lenguaje que usan los check-ins.

const SAGE = '#A8C39A';
const SAGE_DEEP = '#7CA39D';

type Props = { size?: number };

export default function CleanSlateFigure({ size = 104 }: Props) {
  return (
    <Svg width={size} height={size} viewBox="0 0 240 240">
      <G transform="translate(120, 120)">
        {[0, 120, 240].map((deg) => (
          <Circle
            key={deg}
            cx="0"
            cy="0"
            r="70"
            fill="none"
            stroke={SAGE}
            strokeWidth="20"
            strokeLinecap="round"
            strokeDasharray="115 325"
            transform={`rotate(${deg})`}
          />
        ))}

        {/* Las bolas en reposo: dentro del anillo, no disparadas hacia afuera
            como en el logo — nada está girando. */}
        <Circle cx="0" cy="-26" r="11" fill={SAGE_DEEP} />
        <Circle cx="22" cy="6" r="6" fill={SAGE_DEEP} opacity={0.55} />
      </G>
    </Svg>
  );
}
