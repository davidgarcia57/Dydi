import React from 'react';
import Svg, { G, Circle } from 'react-native-svg';

// "Solo en el squad": los tres arcos del logo, pero dos apagados a hairline y
// solo uno vivo. Un squad completo es el anillo cerrado; tú eres un segmento.
//
// Constantes tomadas de DydiLogo para que sea la misma familia visual: r=70 en un
// viewBox de 240, trazo 20 (8.3% del diámetro — el ratio que hay que conservar a
// cualquier escala), cap redondo y dasharray "115 325" para que cada arco cubra
// un tercio. Los arcos van a 0/120/240 grados.
//
// El arco vivo va desplazado 12 unidades: es la firma de Dydi (un elemento fuera
// de registro) y cae temáticamente bien en una app sobre fallar un día.
//
// Paleta decorativa a propósito: terracotta + hairline. Sage/amber/coral son
// semánticos (cumplió/pendiente/falló) y usarlos aquí erosionaría ese lenguaje.

const HAIRLINE = '#E7DECD';
const TERRACOTTA = '#C26F4D';
const INK = '#2A251F';
const INK_SOFT = '#6F6557';

type Props = { size?: number };

export default function SoloSquadFigure({ size = 96 }: Props) {
  return (
    <Svg width={size} height={size} viewBox="0 0 240 240">
      <G transform="translate(120, 120)">
        {/* Los dos lugares que faltan: presentes pero apagados. */}
        <Circle
          cx="0"
          cy="0"
          r="70"
          fill="none"
          stroke={HAIRLINE}
          strokeWidth="20"
          strokeLinecap="round"
          strokeDasharray="115 325"
          transform="rotate(120)"
        />
        <Circle
          cx="0"
          cy="0"
          r="70"
          fill="none"
          stroke={HAIRLINE}
          strokeWidth="20"
          strokeLinecap="round"
          strokeDasharray="115 325"
          transform="rotate(240)"
        />

        {/* Tú: el único segmento vivo, y el que va fuera de registro. 8 y no 12
            como el logo: a este tamaño 12 leía como anillo mal armado en vez de
            desplazamiento deliberado. */}
        <G transform="translate(0, -8)">
          <Circle
            cx="0"
            cy="0"
            r="70"
            fill="none"
            stroke={TERRACOTTA}
            strokeWidth="20"
            strokeLinecap="round"
            strokeDasharray="115 325"
          />
        </G>

        {/* Las dos bolas, en la posición del logo pero más grandes: el logo vive a
            32px y esto a ~104px, donde r=7/r=4 se volvían motas invisibles. */}
        <Circle cx="82" cy="-40" r="11" fill={INK} />
        <Circle cx="97" cy="-19" r="6" fill={INK_SOFT} />
      </G>
    </Svg>
  );
}
