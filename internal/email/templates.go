// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package email

import (
	"fmt"
	"time"

	"github.com/Frimurare/WulfVault/internal/database"
	"github.com/Frimurare/WulfVault/internal/models"
)

// ---------------------------------------------------------------------------
// Prudencia brand constants
// ---------------------------------------------------------------------------

const (
	brandPrimary   = "#004155" // Dark teal — headers, CTA buttons
	brandSecondary = "#5f828c" // Muted teal — accents, borders, secondary text
	brandDark      = "#323e48" // Dark charcoal — card backgrounds, footer
	brandWhite     = "#ffffff"
	brandLightBg   = "#f4f7f8" // Very light teal-grey for page background
	brandCardBg    = "#ffffff"
	brandTextDark  = "#1a2a32"
	brandTextMuted = "#5f828c"
	brandSuccess   = "#0d9488" // Teal-green for success states
	brandWarning   = "#d97706" // Amber for warnings
	brandDanger    = "#b91c1c" // Red for danger/deletion
)

// Prudencia logotype as inline base64 PNG (500x293, white on transparent).
// Works in Gmail, Apple Mail, Outlook.com, and most modern email clients.
// Outlook desktop may block external images but inline base64 usually renders.
const prudenciaLogoPNG = `iVBORw0KGgoAAAANSUhEUgAAAfQAAAElCAYAAAAStBAAAAAAGXRFWHRTb2Z0d2FyZQBBZG9iZSBJbWFnZVJlYWR5ccllPAAAILlJREFUeNrs3U9oJOl5x/FnhslhQ6B7yWUNAbW9vvikXmIMATvqwUcH1HNaHwJqxT7EBzMackgCMeohf0xIzGjAhxzWmdYp+GBGAznkYrblbHAOG0YihxCwsy0I2JewLQj4YINSD3or09tTVe9bXfVWvW/19wPF7kgtqbv+/d7nrbfeunNzcyMAACBud1kFAADE7x6rAMg1MQvicJEsR6wGEOgA1g2SZY/VACAGdLkDAECgAwAAAh0AABDoAACAQAcAgEAHAAAEOtAF95PlTgTLY8vn8Pm33zHrSd/DObsM4A/3oQPw6cL8d27+20+WsdxO2MM9/gAVOoBILZNlliyjZPl0spyySgACHUDY+pbvL0ylrl3yl6wugEAHEJaB3Haxa2C7zK2urx0my1NWHUCgAwiDBrheN9fr471keWICu+/4s4fJcs1qBAh0AO3om+B+YoJ81Z6p1kcOv2dmXkeoAyUxyh1o10SKH9GaVrwhG5ow7xW8Rr/3vqnAZ5bfd2FC3fY7ARDoQDAGUnz7Vj/w96+NkWclXv/MhPXEIdT1Nc/ZRQA3dLkDaCrMUwcOVbo6E/ukOAAIdAAthHnZUJ8Kt7QBBDqA1sJcb0HT+8uvLaE+dfx7AAh0ADXSAXAnBd/XANeBbzqYby6318uvCl5/LG7X05lRDiDQAdREB+jpde1eQZhrgM/WwlgbAUXd5tpAGFj+9pTVDxDoAOqhobpjCfOsW+yW5nt5od4zDYUiC6p0gEAHUJ0G8sMNwnw11MeSf019V+zTxJ6xGQACHUD16jzPRNwmv1lI8Wxx+jf6lkBnBjmAQAdQoTrPm/zmacnKWYP/Uc73elTpAIEOwJ9JztevCir3gdwOhsuig+DOS/6t1JzNARDoADYzzvm6hvky42uLZPkoWV4my01OsOc1BHYK/l5a4QMg0AGUNJLs29T0WvZs5d99E7bH8vpI+MVK1b5aaZ8X/E0CHSDQAdQor9v8LKPi3i147cSE8Wr1PSv5NwEQ6AA2lDfqfLHy/1p5Pyz4HTqg7pmp9E8KGgWrry9yzmYBCHQA9Ziv/P+4xM/trFTgev2d29AAAh1A4FW8y+u5Jg4Q6AAiCvE8exv8zIDVDxDoAOoxrFBlX1gaBbZu+B1WP0CgAyhn4VCh6+C2K8ffpw9YSe9dH1kCHwCBDsBzoK+H8ZHD79LKe7ry7/EGgc4tbQCBDmAD85yv72ZU6Q8kv7v80jQCFisV/kHJv6kGbBIg2z1WAQCLS8meNEYr7NlaqGvgTkwlPTDV9lxev+f8qKCKP6NCBwh0AH6q9KxAP5LXZ3vT6+Mnlt83KAh029PURmwOIBtd7gBsZjlf390wYDXweznfmxb8nHbT77E5AAIdwGa02/xqgwCWnKp+P+d7OgJ+UfCzYzYFQKADqCYvuPdKBK1e/36S871rh8bBhM0AEOgAqjmT/BHsM7HPHDeQ4tHrJ5bqfCB0twMEOoDKdLBb3kC2nuRfZxcT9meSf9380qE6n7IJAAIdQD1mJnyz7OcEfl/yR8mLqfptXfZanR+w+gECHUB9imaE0+vjw4xAHxT8zESKu9rVCasdINAB1Eur7aeW76+Guob1SLKvvx+K/b7zseSPigdAoAOoWKXndb2n19PXn3k+XPuZp1J83T2t7mesboBAB+DPWPJHve+aSr2fUalrqJ+K28NcigbSASDQAdRgIcWD2bJCfWlCfeLw+6fCbWoAgQ6gERrYhxuEuo0G/jGrFyDQATRnJsWD5LJCvYhea3/GagUIdADN0+vhpzWE+lCKZ5MDQKAD8GziEOrpaPeiMGcQHECgAwg81Hfk9fvUCXOAQAcQYahraL+UVyPdCXOgJvdYBQA8hLoqmn/9mQnzCWEOUKEDCDvUH1le85AwBwh0AOHTh6ocshoAAh1AM4Yef/eMUAcIdADNmBPqAIEOIH49Qh0g0AEQ6mVD/ZrVDRDoAMo7cwzRpkJ9RKgDBDqA8i5MiF4FEuoXhDpAoAPYPEQ1pC8JdYBABxC3pQlRQh0g0AFsYaiPCXWAQAcQf6g/l1fzsxPqAIEOINJQV88IdYBAB7A5n9ewCXWAQAfQkCeeQ5RQBwh0AA3xHaKhhvolmx4g0AFCnVAHCHQAjTkpGaJHWxTqZd8PQKADaE3Z0NJr6jNCHQCBDsQf6geEOgACHSDUCXWAQAfQQGidE+qF74db2gACHYgm1E8DCvVhiffjO9R1EGGP3QQg0IFYTAIK9bLvx1eoz8xnBUCgA7Uaev79ZUP9LFn6HQ11whwg0AFvfE/LWjZE9+X2caddC3XCHCDQAe98XzMuG6K7HQt1whwg0IHOhfrTiEP9hDAHCHSgLWWnZZ16fj867ethpKH+sGSolw3zK/E7LS4QvDs3NzesBSBf3wTjruPrTxuq1p85vlYbJCO5vf3MF5fwvTbv48JDmDfxGQEqdCByoU32kgZeTJU6YQ4Q6EBQoR7SfeFlQ13D1OdtdkWhTpgDBDoQFFslGnKo75hKvelQPyTMAQIdINTdQ91lfvNew6F+6Pj5CXOgBgyKAzYPUtcQamKg3NCEtcs852WuaVd5P1TmAIEOEOodCHXCHCDQgU6GehOBFEuoE+ZAzbiGDlQzkbCmZb0Q92eGN3FNnTAHCHSAUG8o1F+K/8sBhDlAoAPRhHpIc62noX7l+Pom5qMvG+anhDngjmvoQP3BHtK0rGWnrnW91ayJMJ+wOwFU6EBbZhLWtKzpLHdlHjIz8bBOCHOAQAcI9YhDnTAHCHRg60J90LFQJ8wBAh3YylD3/QCVTUL9qMJnJ8wBAh3YylBv4r7wso+DfSLlB8kR5gCBDhDqDYa6j4fMEOYAgQ4Q6g2Gukj9T44jzAECHdiaUH8gYU3LWleoE+YAgQ5slTMJb671qqF+UjLMnxLmQP2YKQ5oR4hPRStTZacVti7PSvwNXzPRAQQ6gQ5EE+pHDYShVtsPHV+rI+X3CHOAQAdQLtSbCsWyVTdhDgSAa+hAu8o86lSkuaeiHdb4+whzgEAHCPXIQ50wBwh0gFCPPNQJc4BABwj1yEOdMAcIdIBQl3afX1411AlzoAWMcgfCpM9Hn8vtU9hcPJLbW858chmRT5gDBDqAiqHexHSqRaFOmAMEOoCIQ12v+Y/N1wAQ6AAiDHU1Ev/T0gIg0AFC3WOoC2EOEOgA4g91AIHgtjUgHkspd0tb0fPLARDoAAh1AAQ6gLpC/UWJUD+T2y57AB3FNXQgbjMT2C4uTUNgyWoDqNABhGUit4PfXOhgujmVOkCgAyDUARDoAAh1AAQ6AEIdAIEOEOqEOkCgAyDUAQSN29aA7pqJ+y1t+sQ0nZt9wWoDqNABhFepP3Z8rQY596cDVOgAAg/2ZwXfZ8IZgAodQARmyXJImAMEOoBuhjphDhDoACIPdcIc6BiuoQPbZyy3t6oR5gCBDgAAQkKXOwAABDoAACDQAQAAgQ4AAAh0AAAIdAAAQKADAAACHQAAEOgAABDoAACAQAcAAAQ6AAAg0AEAINABAACBDgAACHQAAECgAwBAoAMAAAIdAAAQ6AAAgEAHAIBABwAABDoAACDQAQAAgQ4AAIEOAAAIdAAAQKADABCpQRt/9M7NzQ2rHgC6a5TxtYtkWbJqvOib9Ttseh3fc9z4dZi3uHKHHn7v0my0EA/Wpt+jbR3XdfIYmr/VlX3TxzZvSpVtOnCoYNo8vmzvL9QwHJllaN7/rsPPXCbLwhwD80DOaU2dT3w5SpadZBkny6ztCt13yX5uNki6Ay0b2Mnf9/j7r9YOiKbD4cayrkcNnUiK1vH9mtaL/o49j58jPbldrJzcQjxxhNCtVmWbTpPl2OF177QUMLb3d1/CaQSOV5ZeDb/vOlnOTBC19RmbOp/4sjTb4koa7npv4xq6npAfJsvzZPnY7DwjideO+UzHZidcmoNhKIiNVjT7K9vyYxMoR9LSNbEtdyZ+emRi1zeNjoU5jx7UFOZifs+B2f8XZt9nG7ibrGyLHfPvTgf6un2z81xEHuzrB8RL04ok2OMP+SfJ8pHZRyeskkYby1NWQ2aQH5v143v9PzF/j+3gZmr5d+cDffXE+X7HWuV7JthPaOV2JtyfyW0vzJRt2oiHHWnoVzU2DcrjGqvxMkXKsQn2MZuisDrfyWgUNVYEhHjb2r68GiHYpZOSVusD9vlOSE9wF5zgGjHb4sZT3xQ5zxuoyF0q9ufCpRDX6rzxKj3U+9B3pHvd1bsdbKhsO05wza3n6RZ+7qE5D+5X/D062PN8Zbmk6GqkOl/df0dNvIl7JV+fjujeZMcs203UW6lqfY40vtzg9/fF7ZaQvM80kjBuD+mS6w3X6Sb7Zt4JbhzAdt30GC2r6dH/D03Dab4l+/PIfN7eBsdBup4uLPvjwOz/I7PvlukBSIsuHTQ34/Rj7VafNhHqZQN9VqGlnN5bOC6x8/TE/yj4owonieHKZ3JtRffMehwJEzvUqeqgyuHKMtqgwbYTSGOtyjEaupm0MFlHS+HwrOTPnJv1UyZcF2Y5M+fBofmv6y1wvZX3uc2hrse87XbaPfM6rw3SJrvclystOm0ZHppqQhxWxFHAITIzB8CbyfLYtJBtdoVRo6Fuy/TE9qbZR8t0T6Y9MHRF+rENXe/jkmGu++d9ExazGo6BiTk/u57LxLzfyRbvl9OaXxdFoOe1tk8dV0To1yjTkc96MLxweD2jd8Pfnuk++o7jfkqo+9fl42ZYIpQ1bB/Jq+vsPs5lQ1P5u4b6Np7PXKrz9Sq9k4Ge7jgTh5NlT+IZTbw07/VRCC021Fa9T0wl5FKxp6HOQLnN2NbxrIPrNh3N3nNcPxoMJ57f08L8nUeOrz+T7buTp+w5fNLlQF/9kLaq9iiyDa0H21OHFhuVXDzSytulWu/J9gzg8nE+KNLFrveZuI0reiHNj9PQc9k7Yu+CT8c8bYuBlJ+K+sBno+duYAdx0Q6zG2HrTxsh5w6fG/EFzqHD6xgrsRkNq8eW13Sp613PEy6DarUhqb1/y5a2ycgh1Ldpn582/HNRBXp6zbLIqIMbnYlJ4q2oXEL9WOiF2fS42Yau977jCf40gMa/a6hvw7MPBqba3oS3Kj20iWVsgR7jiXFuqdJ3hBnkuh7qJ6yqjdgCbKcD61bfv+26+aWE05PnMjtibwuq9GnLPx9FoF9YWn+xVjpdbKjg1ba1dQ97H93aUS5d7wcSby+XS5V3HeC+M3fcLl0tVKpU56vrp/bepbuBHsRdMyfQO99aP3d4Dfys25nE2fXusk9oZb4M9L3ftyxdnQDIZYD2eU2/p5R7ka3IWFt8C8v3ub0pfnri/cihSp+zqjZat9rQz+uaTmdfjKlS7ztUeTqiPeRR49u4L/fFfvnjyhzret7fsQT6SZ0Nn7uRrcyYW3xXVOidpgevrRtywmraeN3aqtn9yALdZV84YtMHWZ3bxjxM1/4rBQ3RWrcxgc57R320tV00BsTLdbMtWrdd6nq3BfqpNPOQHZSrzm0BfCWvxkzNxD69+VGd+2yIgb7X0Z1hl7DvPJdbL7lNsVoIXlsqnlkEn2Mg9of/TNncUVfnrtux1io9tEC3nexiHTDX7+jnQnYlWWTEKtrYQrrR9W7bB15QnUdfnUuJKn1CoMfF9rmo0LsVOpdU6F4bTLF3vdsC/YzNHOQ5vGx17lql79QV6ncD28ltoz7nHQ30OcdLpxSdkPWkMGAVVaInv5i73ocV9h+0wxbKVwX7nEuVPq3jTYYS6H2HA/Bc4uyG0pN30TzN10KXe9fYGmjc1VDNQuLuet+1nOfosQuvAWl7cM604vdrqdJDuA+9b06AthU2i3RnsL1vWuPdY2ugDT1ud21AjhoI1LYb1ycmsPcsx94gsIAcVNx3EFd1vnqet03zO62ac20H+kjcHht4JXEG+lTso/aZ57t7bAHi8/rugVSfltLmsYQxCnsi8U04Q6BvX3WenhP0XH9sqdLHVRr7d1vcqfVAe1/cngE8iXBHOLJsPHXOAdxZl5YKHfX0FNjODbFNOLNgswZ3Hq+r2LTNU+Hy94II9JE5qE5MgH1Uoop4KnENGuubFtYTxwoe21mlox56rL2wvGYmTOiDzXKrzvkC0iq9SKUHOW0a6BqwNyUWrcSfJ8tDhxW0XuXEMv1h+lzjhRQPgou1oQKEaiL2Ue+MVUFZtrDe5FKwS5W+caG3aaCPHd5UVZcS/iQcQ9Pg0JPFx3Lbxd5z/GzM0wzUYyn2rvc9jjmUrM5t45+mG+6r3qr0exUOoLGpvH04lXaum48dD/qBuF37j7WhguoGrIJGpV3v+5YTsL5uwepCC9X5apVum0ZWvz9vKtBFXj3k/rjGlXi1UvG2dVLQDbnr6fenYc711e4ravD5DJSrBgIr1ECcmPdmG/UecoOahmD7hp6q8/UqvSg7982+UOpYq3rb2tSxa8LlJDSTmp8NW6FKvxC3rvMyzs3vJsypzn0G4ky2d7Bl2vX+vOA1add7W7eLLiruO/CvzpHtVar0qZTsqa5jlPum19P1Z7Rr/YHZiaeBhN1C6u3u18/5iMp861r4MVa4XeAy6n3aYnDatv2ITdh6Y/zAYf+po/Fpa1QelN1P79b0xoru87w01aku2kV/mCzvyO2o8ImEOfpU39PTGoL81GwQJo/ZLraTMnMP+DWRsOd6P7f0IKA9trCuc5Kz2ke813Uf+jxZ7uQsQ3OCG8mrqe1iOKEdSfHkIEUH66EJ8glVOYFOoDcu9FHvtu3PE/m6XZ2v7qe2xkGpKv0u29B6YLlcTtDgv28aMCOzkQjy7T0pFA2qfMEqaoT2sp06nJwHLby3OYG+9dX5apVuM3H9ZQR6sYXjyhy0GODnAawnW0W6TY0bW9XHBCfNbouix1a21fVuC/TS105RWd+hITX18HcXDg3PI3Gc6ZBAd2vp266npycGppfMdrFFJ4WJw/6EZoTa9a7vy9ZTM2HzNd746zVcnbs2FHqu+yiB7r6xbdfTtZs1tMFvNDDCOim8EC7FtFEN2xrk0xYq4jOHfYkqvbnz5JHDPuJLbVU6ge7O5Xr6QQst64WlkdGEotu0rjkp/D/udmjHVMLrep+JfSQ++0v3q/Naq3QCvVxwuoT1iTT7eMxFhbBtItC3pbt9ZjkpnAsP42lLqF3vtsDW2cJGATdgz8w+nbfEMHd+29V52SqdQK+R6/X0M2muu9sWEr5PCAMpnuZ0GwJ9LPYn7E05fFo1dzh2nzTcGHe5D7nJc0nZ975vGkJ5SwzjRSYBVOdlqvQJgV4vl+vpOw3uBLbA9B3oo4rvL3ZDh239guo8CFMp7noXabbr3WW2sJ7Zd/qBrUfb/dpPJY4ZEUOozstU6VMC3U9FZmtZ70szXU5LSwMjneS/rQOiy6O6027Hohb+tfDYzlC4dL3vNry9phLXgFtdf7YHcl1JHD1SEynuXWyyOndtQOwU7cME+uYtqYnD65rqwps77Li+qvOigXeX0t1R3UPT+2B7jO6RMHd7SObidtkspCpRTEU8a3nd6XnkmePrYjjupxW/H1yVTqBvznW+9yaugdkO9GNPVfq04vuK1dgEgy3MTzu8DmI2FXvXe9ONjMeOoT6XdrrfTxzD/KnEcXkpxOq8cpVOoFdvWYdwPf3C4X3MPOx0RQ+SuO5gmPXNZ3ruUMVdCpODhMr2QKm2TuIusz7umeN92OA+rwH90OG1lxLP5aUQq/MyVTqB7rFaC+F6+onDiaCugJ2I/TramXSnu71vDvCF2AcDpSe2EYdG0C4cq+KmzyUuD4TSIuFlA6EzNvv8Xsf2+VHA1blrAbaXtb4J9HpaUy6VmO/r6TOHk0F6Ha5Kl502TGxdb10YCJbO7awNk49NA8bl2mp6YmNGuDiqtMuA3s/S7DuukzEdlzj/lA08rcqfO+7z1xLXkyVDrs5Tuv7Py75PAr0eoVxPdx1cc7FBa3podrInjgdMTIHWN+tjYno65ibE9YS2X+L3EObxmQT2ftJQd21o7JgG9kJKPMSjYF3ovv++uD+X/dq831huTx1ZPlsI1blrw+K1Kv0ex3OtYaord9dy8J2Jv66puWlYPHQ4CbxvThoz83NZB+RgJehcD/BzaecWm7TB4Wog9kFtZbwIpEqZSDNdn0fSjTkG0q7348DeU1olu07fvGMa20/k1bwHF5ZjYmgW/VtjKT+6P8YGrC0kQ5puN63S9yyfZ0Sg+zE2B1HP0qqair9uHZeGRWp3reJObzPry2bzwF9Je4ONeiUaHXW6NtsylBPBTs0NlaJeja6Ymv12N6D3lFbqul8dlPzZfXm9Z2n1FtI6jpNQGrB1VuchDuSdmuKrKE+0OFnoP+hyr9dC3Lrwjj1XUSPZ7NrgrtlBNjmxXZuT4jZ1N5+bCoeHaMRvEuB7SifCeSTVH3KUHtt1hPmjSI912zY+CfAzpVW6U68DgV6/EK6nl70OV9WVxHUdrY4gv28+84JdvhNCHPW+GjRDcbutzfd+/06kDVitYg8sBUmon2tq+f6B+XwEuicu96enD3Hx2bofOTYuqnghr2ZN67rTlSCfs5t3zlTCGvW+amH2u/vS/KQ4+vcOI2+020IxxOq8dJVOoPvjcn/6nvi9RWJpGhc+TgL6+x5It7vZr02DRU9mb8qrUcDorkng729uqrEHDVTsl2bf1783i3ibxlydl6rSCXS/LWqXk4Pv6+mrJ4HDGk4Cqwd5lx68cmnWzWPz+bRrMb0XfSbcirYtQu56X5XeLfNpue2Fq6tn4cr8Pt3/XZ4kSHUeTpV+dOfm5mb9iyNLSC0i25h9KZ7Q5cLzxhyK/Vp50+t1YIIqvW1l13KAp7e/nLW8/V3WZZnei1i7D0cBvIcqx81Aip8tMA98Hfs+Z1Q5rkcrx7UutlvRzs0xnR7jIRwTdZ+zbeeNULdn2eNmmRXo2F6jCHdyABzbSBDoAAB0ANfQAQAg0AEAAIEOAAAIdAAAQKADAECgAwAAAh0AABDoAACAQAcAgEAHAAAEOgAAINABAACBDgAAgQ4AAAh0AABAoAMAAAIdAAACHQAAEOgAAIBABwAABDoAAAQ6AAAg0AEAAIEOAAAIdAAACHQAAECgAwAAAh0AABDoAAAQ6AAAgEAHAAAEOgAAINABACDQAQAAgQ4AAFpxj1UABzvJ8ofJ8jsrX/s4Wb6TLB/U/Le+mCyfL/kzz5PlyuPnP0qWccbXz5LlxPN6f5Dx9Z8ly/czvv5usnwq4+t1vMe8313339nkPVTZ/i6fS0psA9f9KcuHOcdT3n7QxDpHRO7c3NywFlAUKD9MlrcLXrNMlsc1nljmybJX8mceeTix6Wc/TZbf1eOk4HV6AP1Xskw8NG70xP8kZ52/mfF1bWT1s47zGt5L3u/O8otk+c9k+WbN6yTvPVTZ/mU+l8s2cPHLnGLqp8ny2ZwGS1aD8lfJ8mucppCiyx1FlctHljAXczLU0HmvQ5/92+az7zmE4R2zjn5kTrwQeSNZhmadvMfqeM1pztc/YxqS634v5/X/wqoEgQ6X6vQfSlZ2X5P8rsSYvEyWP9mgqr1jqqifsPt8Yp18jVB/zddNdZ21vr6V0bDOqua1Z+iAVQkCHTY/lM26af8q8s89N5VlFW8T6q/5g5zKc5v9Y87Xv7L277/MeZ32flyxGkGgwyWUsuh1w3Pz3yza1frFSD+z9i7Yrt0vVxbb+qP7/ZOV5ymr4bX9LWsA01trx9Bncn7+z1iFINDhcqLJol2EOghoZP77q5zXfd7T+3pkgiFrqWNA3N8UfO/nyfIl87nT5Uvm63n2t6QqTbeBro/vFewXXwj0/b+5ti+d57zufO11b1b8u1emys7yR+a/70l2T5kOnvtAgDXctgZXv1z799/KJ29jS30Y4Wd7r+BY0FvTsm4Z0hPqpyR/BHJalY62ZP/4wCwzE1TrQfQGh9Br9Br4RxnrKh0E95Wcn5uw6kCgw4UGVNatUnpC1mvDXzbVxZ926DPnnTiXkn//b0q/n3fr0xe2cP/RUL+WzW4F2zZ6HF3K6+M29Lysd1q8RXUOAh1VTzLLnBPy26ai+JGpLpoclKMD7o5zvle1+/OtnK8/dvx5nXTnG+w62IDeq//PGV//45zXf5dVBgIdZTzOqdKVdg/q4LGFqRYmDVUMb4ifbtt3c76u14Jdr81/XzafNaxrdNxAL2d94nUfmOPo7YzjbN1SmBkOBRgUhyx60view+vSCVXmEX/WvGk//5fdoDTtJv6PnDDiFqt8kxINbYBAR2k6+cVXxX6LVlqxfyzca7yNblYWnZDnjYqhta1V+s8trynTYwQCHXiNdiPr9Wm9ZeynltfqNfezCD9j3qj832Dz1+ZMGMhl89eW73MfP6y4hg4XJ2axPbBER+t+0dPJWxsU/+2pOso7Nlw/C0/DKg7zBxxCTsfYsWQPRtXq/OusIhDoKCvv8aXpIypHcjuQLG+u94mnQP+ux3DMG9U/k+ynX63LeyJd092kofQqaPe73o71TaEyL0PXWdZshTyEBQQ6NvIXOSeVvZVKS7viv5Hzus9G+Jl/ILcPEVmnIf2epTp6T/Knyq16Iv5Zztd7Ob0EbRzPNxkNO/33PxHmQLO4hg5Xo7V//1aHPtufS/a82mKCPm9e9uc5DYE06Ko+DSvvVrg7Ge8pryfA9+1iedd+9T7qdzlsAAId7flxztf7pmLUud7nBVXpjz29ryfyyRHVq0vVx7bqpYS/L/i+Tu2qU9++NJ/9pfn3uOBn6noa1rLgPf3EvJ+fFLwX37eL6YyBP81pdDwT7nwACHS05u8KqtW3TLDuWX4+Rl+X4pH82p09NJ99KMXd20upbw73HxR8723zft4ueE0TM4t9OacnQG9h+1cOKYBARzuKngJlcyFxTyDyWbHfnmfzC6n+TPX1hsYvNvxZ/SwnDe0zv1/QCJxzWAEEOtox2iBE9PXjDnx2DfXzCgH6OQ+Nms9tsD1+Ls0OUPx+wXrTXoRvc1gBBDra8evmBH3jWJn7CLI2GzRfFfvsXSntYn9kAtTHOrgy6/fC4bU3Zrt9qqX1lnfNn0FygGd3bm5uWAsoooOavpUsv50sg5Wv/0+y/HuyfEfqvT0p7z74Is89Nib0/Uwa/Pyu20PD8zdXvr5Iln+T2xH7PtbFuzmNhJOM95c3kcyHFddV3nuoc/vn7X8fNrCd2/zb6ID/E2AAEl7SG3Z8HBsAAAAASUVORK5CYII=`

// prudenciaShieldImg returns an <img> tag with the Prudencia logotype
func prudenciaShieldImg() string {
	return `<img src="data:image/png;base64,` + prudenciaLogoPNG + `" alt="Prudencia Security" width="200" height="117" style="display:block;margin:0 auto 15px auto;" />`
}

// emailHeader returns the standard Prudencia email header
func emailHeader(title, subtitle string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"></head>
<body style="margin:0;padding:0;font-family:Arial,Helvetica,sans-serif;background-color:%s;">
<table width="100%%" cellpadding="0" cellspacing="0" style="background-color:%s;padding:30px 0;">
<tr><td align="center">
<table width="600" cellpadding="0" cellspacing="0" style="background-color:%s;border-radius:12px;overflow:hidden;box-shadow:0 8px 30px rgba(0,0,0,0.12);">
<!-- Header -->
<tr><td style="background-color:%s;padding:35px 30px 25px 30px;text-align:center;">
%s
<h1 style="color:%s;margin:0;font-size:22px;font-weight:700;letter-spacing:0.5px;">%s</h1>
<p style="color:%s;margin:8px 0 0 0;font-size:13px;letter-spacing:0.5px;">%s</p>
</td></tr>`,
		brandLightBg, brandLightBg, brandCardBg,
		brandPrimary, prudenciaShieldImg(),
		brandWhite, title,
		brandSecondary, subtitle)
}

// emailFooter returns the standard Prudencia email footer
func emailFooter() string {
	return fmt.Sprintf(`<!-- Footer -->
<tr><td style="background-color:%s;padding:25px 30px;text-align:center;">
<p style="margin:0 0 5px 0;color:%s;font-size:11px;letter-spacing:0.5px;">Prudencia Security &middot; Secure File Transfer</p>
<p style="margin:0;color:%s;font-size:10px;opacity:0.7;">This is an automated message. Do not reply to this email.</p>
</td></tr>
</table></td></tr></table></body></html>`, brandDark, brandSecondary, brandSecondary)
}

// ctaButton returns a styled CTA button
func ctaButton(text, href, bgColor string) string {
	if bgColor == "" {
		bgColor = brandPrimary
	}
	return fmt.Sprintf(`<table width="100%%" cellpadding="0" cellspacing="0" style="margin:25px 0;">
<tr><td align="center">
<a href="%s" style="display:inline-block;background-color:%s;color:#ffffff;padding:14px 40px;text-decoration:none;border-radius:8px;font-size:15px;font-weight:600;letter-spacing:0.5px;box-shadow:0 4px 12px rgba(0,65,85,0.3);">%s</a>
</td></tr></table>`, href, bgColor, text)
}

// infoRow returns a key-value row for file info tables
func infoRow(label, value string) string {
	return fmt.Sprintf(`<tr>
<td style="padding:10px 15px;color:%s;font-size:13px;font-weight:600;white-space:nowrap;border-bottom:1px solid #eef2f3;">%s</td>
<td style="padding:10px 15px;color:%s;font-size:13px;border-bottom:1px solid #eef2f3;">%s</td>
</tr>`, brandSecondary, label, brandTextDark, value)
}

// fileInfoTable returns a styled file info card
func fileInfoTable(rows string) string {
	return fmt.Sprintf(`<div style="background:%s;border:1px solid #dce4e6;border-left:4px solid %s;border-radius:8px;overflow:hidden;margin:20px 0;">
<table width="100%%" cellpadding="0" cellspacing="0">%s</table>
</div>`, brandLightBg, brandSecondary, rows)
}

// statusBanner returns a colored banner for status messages
func statusBanner(text, bgColor, textColor, borderColor string) string {
	return fmt.Sprintf(`<div style="background-color:%s;border-left:4px solid %s;padding:16px 20px;margin:20px 0;border-radius:6px;">
<p style="margin:0;color:%s;font-size:14px;">%s</p>
</div>`, bgColor, borderColor, textColor, text)
}

// ---------------------------------------------------------------------------
// Upload notification (file uploaded via file request)
// ---------------------------------------------------------------------------

func GenerateUploadNotificationHTML(request *models.FileRequest, file *database.FileInfo, uploaderIP, serverURL string) string {
	uploadTime := time.Unix(file.UploadDate, 0).Format("2006-01-02 15:04:05")

	return emailHeader("New File Uploaded", "Upload Notification") +
		fmt.Sprintf(`<tr><td style="padding:30px;">
%s
%s
%s
<p style="text-align:center;color:%s;font-size:12px;margin-top:5px;">Or go directly to: <code style="background:#eef2f3;padding:2px 6px;border-radius:3px;font-size:11px;">%s/dashboard</code></p>
</td></tr>`,
			statusBanner(
				fmt.Sprintf(`<strong>Good news!</strong> Someone has uploaded a file via your request: <strong>%s</strong>`, request.Title),
				"#e6f5f0", brandSuccess, brandSuccess,
			),
			fileInfoTable(
				infoRow("Filename", file.Name)+
					infoRow("Size", file.Size)+
					infoRow("Uploaded", uploadTime)+
					infoRow("IP Address", uploaderIP),
			),
			ctaButton("VIEW IN DASHBOARD", serverURL+"/dashboard", ""),
			brandTextMuted, serverURL,
		) +
		emailFooter()
}

func GenerateUploadNotificationText(request *models.FileRequest, file *database.FileInfo, uploaderIP, serverURL string) string {
	uploadTime := time.Unix(file.UploadDate, 0).Format("2006-01-02 15:04:05")
	return fmt.Sprintf(`New File Uploaded

Someone has uploaded a file via your upload request:

Request: %s
Filename: %s
Size: %s
Uploaded: %s
IP Address: %s

Log in to view and download the file:
%s/dashboard

---
Prudencia Security — Secure File Transfer
`, request.Title, file.Name, file.Size, uploadTime, uploaderIP, serverURL)
}

// ---------------------------------------------------------------------------
// Download notification (someone downloaded your file)
// ---------------------------------------------------------------------------

func GenerateDownloadNotificationHTML(file *database.FileInfo, downloaderIP, serverURL string) string {
	downloadTime := time.Now().Format("2006-01-02 15:04:05")

	return emailHeader("File Downloaded", "Download Notification") +
		fmt.Sprintf(`<tr><td style="padding:30px;">
%s
%s
%s
</td></tr>`,
			statusBanner(
				`<strong>Good news!</strong> Someone has downloaded your file. Here are the details:`,
				"#e6f5f0", brandSuccess, brandSuccess,
			),
			fileInfoTable(
				infoRow("Filename", file.Name)+
					infoRow("Size", file.Size)+
					infoRow("Downloaded", downloadTime)+
					infoRow("IP Address", downloaderIP)+
					infoRow("Downloads remaining", getDownloadsRemainingText(file)),
			),
			ctaButton("VIEW IN DASHBOARD", serverURL+"/dashboard", ""),
		) +
		emailFooter()
}

func GenerateDownloadNotificationText(file *database.FileInfo, downloaderIP, serverURL string) string {
	downloadTime := time.Now().Format("2006-01-02 15:04:05")
	return fmt.Sprintf(`Your file has been downloaded!

Someone has downloaded one of your files:

Filename: %s
Size: %s
Downloaded: %s
IP Address: %s
Downloads remaining: %s

Log in to see details:
%s/dashboard

---
Prudencia Security — Secure File Transfer
`, file.Name, file.Size, downloadTime, downloaderIP, getDownloadsRemainingText(file), serverURL)
}

// ---------------------------------------------------------------------------
// Splash link email (file shared with someone — THE main customer-facing email)
// ---------------------------------------------------------------------------

func GenerateSplashLinkHTML(splashLink string, file *database.FileInfo, message, senderEmail string) string {
	senderBlock := ""
	if senderEmail != "" {
		senderBlock = fmt.Sprintf(`<div style="background:%s;border-radius:8px;padding:14px 20px;margin:0 0 20px 0;">
<p style="margin:0;color:%s;font-size:13px;"><strong style="color:%s;">Sent by:</strong> %s</p>
</div>`, brandLightBg, brandTextDark, brandSecondary, senderEmail)
	}

	messageBlock := ""
	if message != "" {
		messageBlock = fmt.Sprintf(`<div style="background:#fefce8;border-left:4px solid %s;padding:16px 20px;margin:0 0 20px 0;border-radius:6px;">
<p style="margin:0 0 5px 0;color:%s;font-size:12px;font-weight:600;text-transform:uppercase;letter-spacing:1px;">Message</p>
<p style="margin:0;color:%s;font-size:14px;">%s</p>
</div>`, brandWarning, brandSecondary, brandTextDark, message)
	}

	return emailHeader("A file has been shared with you", "Secure File Transfer") +
		fmt.Sprintf(`<tr><td style="padding:30px;">
%s
%s
%s
%s
<p style="text-align:center;color:%s;font-size:11px;margin-top:15px;">Or copy this link:<br/>
<code style="background:#eef2f3;padding:4px 8px;border-radius:4px;font-size:11px;word-break:break-all;color:%s;">%s</code></p>
</td></tr>`,
			senderBlock,
			messageBlock,
			fileInfoTable(
				infoRow("Filename", file.Name)+
					infoRow("Size", file.Size),
			),
			ctaButton("DOWNLOAD FILE", splashLink, brandSuccess),
			brandTextMuted, brandTextDark, splashLink,
		) +
		emailFooter()
}

func GenerateSplashLinkText(splashLink string, file *database.FileInfo, message, senderEmail string) string {
	senderLine := ""
	if senderEmail != "" {
		senderLine = "Sent by: " + senderEmail + "\n"
	}
	return fmt.Sprintf(`A file has been shared with you

%s%s
Filename: %s
Size: %s

Download: %s

---
Prudencia Security — Secure File Transfer
`, senderLine, getMessageText(message), file.Name, file.Size, splashLink)
}

// ---------------------------------------------------------------------------
// Account deletion (GDPR)
// ---------------------------------------------------------------------------

func GenerateAccountDeletionHTML(accountName string) string {
	return emailHeader("Account Deleted", "GDPR Compliance") +
		fmt.Sprintf(`<tr><td style="padding:30px;">
<div style="text-align:center;margin:0 0 20px 0;">
<div style="width:60px;height:60px;background:#e6f5f0;border-radius:50%%;display:inline-flex;align-items:center;justify-content:center;font-size:28px;line-height:60px;">&#10003;</div>
</div>
<p style="color:%s;font-size:15px;">Hi %s,</p>
<p style="color:%s;font-size:14px;">This is a confirmation that your download account has been deleted from our system in accordance with GDPR.</p>
%s
<p style="color:%s;font-size:14px;">We respect your right to erasure under GDPR and confirm that all your personal information has been handled in accordance with data protection regulations.</p>
</td></tr>`,
			brandTextDark, accountName,
			brandTextDark,
			statusBanner(
				`<strong>What happened:</strong><br/>
&bull; Your personal information has been permanently anonymized<br/>
&bull; You can no longer download files with this account<br/>
&bull; To download files again you must register a new account`,
				"#fef2f2", brandDanger, brandDanger,
			),
			brandTextDark,
		) +
		emailFooter()
}

func GenerateAccountDeletionText(accountName string) string {
	return fmt.Sprintf(`Account Deleted — GDPR Compliance

Hi %s,

This is a confirmation that your download account has been deleted from our system in accordance with GDPR.

What happened:
- Your personal information has been permanently anonymized
- You can no longer download files with this account
- To download files again you must register a new account

We respect your right to erasure under GDPR.

---
Prudencia Security — Secure File Transfer
`, accountName)
}

// ---------------------------------------------------------------------------
// Welcome email (new user account)
// ---------------------------------------------------------------------------

func SendWelcomeEmail(email, resetToken, serverURL, companyName, adminName, adminEmail string) error {
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", serverURL, resetToken)
	subject := fmt.Sprintf("Welcome to %s — Set Your Password", companyName)

	htmlBody := emailHeader(fmt.Sprintf("Welcome to %s!", companyName), "Your account has been created") +
		fmt.Sprintf(`<tr><td style="padding:30px;">
%s
<p style="color:%s;font-size:14px;">To get started, you need to set your password and log in to your account.</p>
<div style="background:%s;border:2px solid %s;border-radius:10px;padding:25px;margin:25px 0;text-align:center;">
<h2 style="color:%s;margin:0 0 10px 0;font-size:18px;">Set Your Password</h2>
<p style="color:%s;margin:0 0 20px 0;font-size:14px;">Click the button below to create your password and access your account.</p>
%s
<p style="font-size:12px;color:%s;margin:15px 0 0 0;">This link is valid for 1 hour</p>
</div>
%s
<p style="text-align:center;color:%s;font-size:11px;margin-top:20px;">If the button doesn't work, copy and paste this link:<br/>
<code style="background:#eef2f3;padding:4px 8px;border-radius:4px;font-size:10px;word-break:break-all;">%s</code></p>
</td></tr>`,
			statusBanner(
				fmt.Sprintf(`<strong>Congratulations!</strong> <strong>%s</strong> (%s) has added you to <strong>%s</strong>. You can now share, receive, and request files securely.`, adminName, adminEmail, companyName),
				"#e6f5f0", brandSuccess, brandSuccess,
			),
			brandTextDark,
			brandCardBg, brandSecondary,
			brandPrimary, brandTextDark,
			ctaButton("SET PASSWORD &amp; LOGIN", resetLink, ""),
			brandTextMuted,
			fileInfoTable(infoRow("Your Login Email", email)),
			brandTextMuted, resetLink,
		) +
		emailFooter()

	textBody := fmt.Sprintf(`Welcome to %s!

%s (%s) has added you to %s. You can now share, receive, and request files securely.

Your login email: %s

Set your password by visiting this link:
%s

This link is valid for 1 hour.

---
Prudencia Security — Secure File Transfer
`, companyName, adminName, adminEmail, companyName, email, resetLink)

	provider, err := GetActiveProvider(database.DB)
	if err != nil {
		return err
	}
	return provider.SendEmail(email, subject, htmlBody, textBody)
}

// ---------------------------------------------------------------------------
// Team invitation email
// ---------------------------------------------------------------------------

func SendTeamInvitationEmail(email, teamName, serverURL, companyName string) error {
	subject := fmt.Sprintf("Welcome to team %s — %s", teamName, companyName)

	htmlBody := emailHeader(fmt.Sprintf("Welcome to Team: %s", teamName), "You've been added to a collaborative team") +
		fmt.Sprintf(`<tr><td style="padding:30px;">
%s
<p style="color:%s;font-size:14px;">As a team member, you can now:</p>
<div style="margin:15px 0 25px 0;padding:0 0 0 10px;">
<p style="color:%s;font-size:14px;margin:8px 0;line-height:1.6;">&#128193; Access all files shared with the team<br/>
&#11014;&#65039; Upload files to share with team members<br/>
&#128101; Collaborate with other team members<br/>
&#128274; Securely transfer files within your team</p>
</div>
%s
%s
<p style="text-align:center;color:%s;font-size:11px;margin-top:20px;">Or go directly to: <code style="background:#eef2f3;padding:2px 6px;border-radius:3px;font-size:10px;">%s/login</code></p>
</td></tr>`,
			statusBanner(
				fmt.Sprintf(`<strong>Congratulations!</strong> You have been added to the team <strong>"%s"</strong> in the <strong>%s</strong> file sharing platform.`, teamName, companyName),
				"#e6f5f0", brandSuccess, brandSuccess,
			),
			brandTextDark,
			brandTextDark,
			fileInfoTable(infoRow("Your Login Email", email)),
			ctaButton("LOG IN TO YOUR TEAM", serverURL+"/login", ""),
			brandTextMuted, serverURL,
		) +
		emailFooter()

	textBody := fmt.Sprintf(`Welcome to team %s — %s

You have been added to the team "%s" in the %s file sharing platform.

As a team member, you can now:
- Access all files shared with the team
- Upload files to share with team members
- Collaborate with other team members
- Securely transfer files within your team

Your login email: %s

Log in here: %s/login

---
Prudencia Security — Secure File Transfer
`, teamName, companyName, teamName, companyName, email, serverURL)

	provider, err := GetActiveProvider(database.DB)
	if err != nil {
		return err
	}
	return provider.SendEmail(email, subject, htmlBody, textBody)
}

// ---------------------------------------------------------------------------
// Password reset email
// ---------------------------------------------------------------------------

func SendPasswordResetEmail(email, resetToken, serverURL string) error {
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", serverURL, resetToken)
	subject := "Password Reset Request"

	htmlBody := emailHeader("Password Reset", "Security Notification") +
		fmt.Sprintf(`<tr><td style="padding:30px;">
%s
<div style="background:%s;border:2px solid %s;border-radius:10px;padding:25px;margin:25px 0;text-align:center;">
<h2 style="color:%s;margin:0 0 10px 0;font-size:18px;">Reset Your Password</h2>
<p style="color:%s;margin:0 0 20px 0;font-size:14px;">Click the button below to create a new password.</p>
%s
<p style="font-size:12px;color:%s;margin:15px 0 0 0;">This link is valid for 1 hour</p>
</div>
%s
<p style="text-align:center;color:%s;font-size:11px;margin-top:20px;">If the button doesn't work, copy and paste this link:<br/>
<code style="background:#eef2f3;padding:4px 8px;border-radius:4px;font-size:10px;word-break:break-all;">%s</code></p>
</td></tr>`,
			statusBanner(
				`A password reset has been requested for your account. If you did not request this, please ignore this email.`,
				"#fefce8", brandWarning, brandWarning,
			),
			brandCardBg, brandSecondary,
			brandPrimary, brandTextDark,
			ctaButton("RESET PASSWORD", resetLink, ""),
			brandTextMuted,
			statusBanner(
				`<strong>Security reminder:</strong><br/>
&bull; Never share this link with anyone<br/>
&bull; We will never ask for your password via email<br/>
&bull; If you didn't request this reset, ignore this email`,
				"#fef2f2", brandDanger, brandDanger,
			),
			brandTextMuted, resetLink,
		) +
		emailFooter()

	textBody := fmt.Sprintf(`Password Reset Request

A password reset has been requested for your account.

Click the link below to reset your password:
%s

This link is valid for 1 hour.

Security reminder:
- Never share this link with anyone
- We will never ask for your password via email
- If you didn't request this reset, ignore this email

---
Prudencia Security — Secure File Transfer
`, resetLink)

	provider, err := GetActiveProvider(database.DB)
	if err != nil {
		return err
	}
	return provider.SendEmail(email, subject, htmlBody, textBody)
}

// ---------------------------------------------------------------------------
// Expiration reminder email (used by cleanup/reminders.go)
// ---------------------------------------------------------------------------

// GenerateExpirationReminderHTML creates a branded expiration reminder email
func GenerateExpirationReminderHTML(fileName, fileSize, splashLink string, daysLeft int, urgent bool) string {
	urgencyText := fmt.Sprintf("Your file will expire in <strong>%d days</strong>.", daysLeft)
	bannerBg := "#fefce8"
	bannerBorder := brandWarning
	bannerText := brandTextDark
	if urgent {
		urgencyText = fmt.Sprintf("<strong>URGENT:</strong> Your file expires <strong>tomorrow</strong>!")
		bannerBg = "#fef2f2"
		bannerBorder = brandDanger
		bannerText = brandDanger
	}

	return emailHeader("File Expiration Reminder", "Action Required") +
		fmt.Sprintf(`<tr><td style="padding:30px;">
%s
%s
%s
<p style="text-align:center;color:%s;font-size:12px;margin-top:10px;">Download or re-share the file before it expires.</p>
</td></tr>`,
			statusBanner(urgencyText, bannerBg, bannerText, bannerBorder),
			fileInfoTable(
				infoRow("Filename", fileName)+
					infoRow("Size", fileSize)+
					infoRow("Expires in", fmt.Sprintf("%d day(s)", daysLeft)),
			),
			ctaButton("VIEW FILE", splashLink, brandPrimary),
			brandTextMuted,
		) +
		emailFooter()
}

// GenerateExpirationReminderText creates a plain text expiration reminder
func GenerateExpirationReminderText(fileName, fileSize, splashLink string, daysLeft int, urgent bool) string {
	prefix := ""
	if urgent {
		prefix = "URGENT: "
	}
	return fmt.Sprintf(`%sFile Expiration Reminder

Your file will expire in %d day(s):

Filename: %s
Size: %s

Download or re-share the file before it expires:
%s

---
Prudencia Security — Secure File Transfer
`, prefix, daysLeft, fileName, fileSize, splashLink)
}

// ---------------------------------------------------------------------------
// Helper functions
// ---------------------------------------------------------------------------

func getDownloadsRemainingText(file *database.FileInfo) string {
	if file.UnlimitedDownloads {
		return "Unlimited"
	}
	if file.DownloadsRemaining <= 0 {
		return "0 (no more downloads available)"
	}
	return fmt.Sprintf("%d", file.DownloadsRemaining)
}

func getMessageText(message string) string {
	if message == "" {
		return ""
	}
	return fmt.Sprintf("Message: %s\n\n", message)
}
