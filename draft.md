# Require package
1. https://github.com/spf13/cobra

# 功能描述
1. 參數中輸入來源的 ASNs
2. 輸入目標 IP，這邊也允許輸入 AWS region，利如 aws_us-west-2，每個 AWS region IP group 所包含的 IP 請參考 http://ec2-reachability.amazonaws.com/，當使用 AWS region 的話，從對應的 IP 中隨機選其中一個即可
3. 計算平均每個 ASN 最多可指派的 probe 數量，由於 Quotas 規定每次最多只能 1000 probes，所以平均每個 ASN 最多可指派的 probe 數量為 1000/#ASN
4. 搜索這些 ASN 的 probe(請參考 Example1)
5. 如果 ASN 可用 probe 數量大於平均每個 ASN 最多可指派的 probe 數量，而從這些 ASN 可用 probe 中亂數且不重複的抽取最多可指派數量。如果 ASN 可用 probe 數量小於或等於平均每個 ASN 最多可指派的 probe 數量，則所有 probe 皆使用。如果有 ASN 沒有 probes 時，需要把所有沒有 probe 的 ASN 以及有 probe 的 ASN list 出來並輸出詢問使用者是否仍需要進行下一步。
6. 透過 RIPE Atlas 創建 measurement，並使用上一步所篩選出來的 probes，measurement 中使用 ICMP traceroute 進行測試。
7. 用 while 這種迴圈方式去呼叫 RIPE Atlas 去檢查 measurement 是否已經測試完成，每次檢測之間間隔 3s，如果檢測到 measurement 完成則進行下一步
8. 將 measurement 結果撈出來並將所有路徑進行比對，以找出 common ASN。


# Quotas
The following quotas apply for users/measurements:

1. Up to 100 simultaneous measurements
2. Up to 1000 probes may be used per measurement
3. Up to 100,000 results can be generated per day
4. Up to 50 measurement results per second per measurement. This is calculated as the spread divided by the number of probes.
5. Up to 1,000,000 credits may be used each day
6. Up to 25 periodic and 25 one-off measurements of the same type running against the same target at any time (targets can opt in to handle more - just let us know!)

# Example
## Example 1
Request URL: https://atlas.ripe.net/api/v2/probes/?status=1&asn_v4__in=5384,7713

Response:
```
{
  "count": 5,
  "next": null,
  "previous": null,
  "results": [
    {
      "address_v4": "125.166.118.254",
      "address_v6": "2001:448a:5130:7c2c:1:81ff:fef4:759c",
      "asn_v4": 7713,
      "asn_v6": 7713,
      "country_code": "ID",
      "description": "Lia Kalibaru",
      "firmware_version": 5080,
      "first_connected": 1630420314,
      "geometry": {
        "type": "Point",
        "coordinates": [
          113.9905,
          -8.2925
        ]
      },
      "id": 50271,
      "is_anchor": false,
      "is_public": true,
      "last_connected": 1759636712,
      "prefix_v4": "125.166.116.0/22",
      "prefix_v6": "2001:448a:5130::/48",
      "status": {
        "id": 1,
        "name": "Connected",
        "since": "2025-09-30T16:41:47Z"
      },
      "status_since": 1759250507,
      "tags": [
        {
          "name": "system: V4",
          "slug": "system-v4"
        },
        {
          "name": "system: IPv6 Stable 1d",
          "slug": "system-ipv6-stable-1d"
        },
        {
          "name": "system: IPv6 Capable",
          "slug": "system-ipv6-capable"
        },
        {
          "name": "system: IPv4 RFC1918",
          "slug": "system-ipv4-rfc1918"
        },
        {
          "name": "system: Resolves A Correctly",
          "slug": "system-resolves-a-correctly"
        },
        {
          "name": "system: IPv4 Capable",
          "slug": "system-ipv4-capable"
        },
        {
          "name": "system: IPv4 Works",
          "slug": "system-ipv4-works"
        },
        {
          "name": "system: IPv6 Works",
          "slug": "system-ipv6-works"
        },
        {
          "name": "system: IPv6 Stable 90d",
          "slug": "system-ipv6-stable-90d"
        },
        {
          "name": "system: Resolves AAAA Correctly",
          "slug": "system-resolves-aaaa-correctly"
        }
      ],
      "total_uptime": 84757515,
      "type": "Probe"
    },
    {
      "address_v4": "36.75.25.184",
      "address_v6": "2001:448a:d020:e89:1:e9ff:fe2f:acea",
      "asn_v4": 7713,
      "asn_v6": 7713,
      "country_code": "ID",
      "description": "Indihome-Kalsel-Rumah Fadloe Robby",
      "firmware_version": 5080,
      "first_connected": 1663580073,
      "geometry": {
        "type": "Point",
        "coordinates": [
          114.6105,
          -3.3115
        ]
      },
      "id": 50462,
      "is_anchor": false,
      "is_public": true,
      "last_connected": 1759636712,
      "prefix_v4": "36.75.16.0/20",
      "prefix_v6": "2001:448a::/32",
      "status": {
        "id": 1,
        "name": "Connected",
        "since": "2025-09-30T23:41:57Z"
      },
      "status_since": 1759275717,
      "tags": [
        {
          "name": "system: V4",
          "slug": "system-v4"
        },
        {
          "name": "system: IPv6 Capable",
          "slug": "system-ipv6-capable"
        },
        {
          "name": "ADSL",
          "slug": "adsl"
        },
        {
          "name": "Home",
          "slug": "home"
        },
        {
          "name": "Fibre",
          "slug": "fibre"
        },
        {
          "name": "IPv4",
          "slug": "ipv4"
        },
        {
          "name": "IPv6",
          "slug": "ipv6"
        },
        {
          "name": "system: IPv4 Capable",
          "slug": "system-ipv4-capable"
        },
        {
          "name": "system: Resolves A Correctly",
          "slug": "system-resolves-a-correctly"
        },
        {
          "name": "system: IPv4 Works",
          "slug": "system-ipv4-works"
        },
        {
          "name": "system: IPv6 Works",
          "slug": "system-ipv6-works"
        },
        {
          "name": "system: Resolves AAAA Correctly",
          "slug": "system-resolves-aaaa-correctly"
        },
        {
          "name": "system: IPv6 Stable 90d",
          "slug": "system-ipv6-stable-90d"
        },
        {
          "name": "system: IPv4 RFC1918",
          "slug": "system-ipv4-rfc1918"
        },
        {
          "name": "system: IPv6 Stable 1d",
          "slug": "system-ipv6-stable-1d"
        }
      ],
      "total_uptime": 94695755,
      "type": "Probe"
    },
    {
      "address_v4": "125.166.118.254",
      "address_v6": "2001:448a:5130:7c2c:da58:d7ff:fe03:80d",
      "asn_v4": 7713,
      "asn_v6": 7713,
      "country_code": "ID",
      "description": "Kalibaru",
      "firmware_version": 5080,
      "first_connected": 1672836929,
      "geometry": {
        "type": "Point",
        "coordinates": [
          113.9695,
          -8.2785
        ]
      },
      "id": 60058,
      "is_anchor": false,
      "is_public": true,
      "last_connected": 1759636712,
      "prefix_v4": "125.166.116.0/22",
      "prefix_v6": "2001:448a:5130::/48",
      "status": {
        "id": 1,
        "name": "Connected",
        "since": "2025-09-30T16:40:56Z"
      },
      "status_since": 1759250456,
      "tags": [
        {
          "name": "system: IPv6 Stable 1d",
          "slug": "system-ipv6-stable-1d"
        },
        {
          "name": "system: IPv6 Capable",
          "slug": "system-ipv6-capable"
        },
        {
          "name": "system: V5",
          "slug": "system-v5"
        },
        {
          "name": "system: IPv4 RFC1918",
          "slug": "system-ipv4-rfc1918"
        },
        {
          "name": "system: Resolves A Correctly",
          "slug": "system-resolves-a-correctly"
        },
        {
          "name": "system: Resolves AAAA Correctly",
          "slug": "system-resolves-aaaa-correctly"
        },
        {
          "name": "system: IPv4 Works",
          "slug": "system-ipv4-works"
        },
        {
          "name": "system: IPv6 Stable 30d",
          "slug": "system-ipv6-stable-30d"
        },
        {
          "name": "system: IPv6 Works",
          "slug": "system-ipv6-works"
        },
        {
          "name": "system: IPv4 Capable",
          "slug": "system-ipv4-capable"
        },
        {
          "name": "system: IPv6 Stable 90d",
          "slug": "system-ipv6-stable-90d"
        }
      ],
      "total_uptime": 84503508,
      "type": "Probe"
    },
    {
      "address_v4": "2.49.28.101",
      "address_v6": null,
      "asn_v4": 5384,
      "asn_v6": null,
      "country_code": "AE",
      "description": "JKUMAR-RIPE",
      "firmware_version": 5080,
      "first_connected": 1720451348,
      "geometry": {
        "type": "Point",
        "coordinates": [
          55.2695,
          25.1805
        ]
      },
      "id": 64056,
      "is_anchor": false,
      "is_public": true,
      "last_connected": 1759636712,
      "prefix_v4": "2.49.16.0/20",
      "prefix_v6": null,
      "status": {
        "id": 1,
        "name": "Connected",
        "since": "2025-10-02T08:54:41Z"
      },
      "status_since": 1759395281,
      "tags": [
        {
          "name": "system: V5",
          "slug": "system-v5"
        },
        {
          "name": "system: IPv4 RFC1918",
          "slug": "system-ipv4-rfc1918"
        },
        {
          "name": "system: Resolves AAAA Correctly",
          "slug": "system-resolves-aaaa-correctly"
        },
        {
          "name": "system: IPv4 Capable",
          "slug": "system-ipv4-capable"
        },
        {
          "name": "system: IPv4 Works",
          "slug": "system-ipv4-works"
        },
        {
          "name": "system: Resolves A Correctly",
          "slug": "system-resolves-a-correctly"
        },
        {
          "name": "system: IPv4 Stable 1d",
          "slug": "system-ipv4-stable-1d"
        }
      ],
      "total_uptime": 38639854,
      "type": "Probe"
    },
    {
      "address_v4": "5.193.37.236",
      "address_v6": "2001:8f8:1471:334e:da58:d7ff:fe03:11e8",
      "asn_v4": 5384,
      "asn_v6": 8966,
      "country_code": "AE",
      "description": "SHJ-65794",
      "firmware_version": 5080,
      "first_connected": 1733828191,
      "geometry": {
        "type": "Point",
        "coordinates": [
          55.3615,
          25.3095
        ]
      },
      "id": 65794,
      "is_anchor": false,
      "is_public": true,
      "last_connected": 1759636712,
      "prefix_v4": "5.193.32.0/21",
      "prefix_v6": "2001:8f8::/32",
      "status": {
        "id": 1,
        "name": "Connected",
        "since": "2025-09-29T23:38:40Z"
      },
      "status_since": 1759189120,
      "tags": [
        {
          "name": "system: V5",
          "slug": "system-v5"
        },
        {
          "name": "system: IPv4 RFC1918",
          "slug": "system-ipv4-rfc1918"
        },
        {
          "name": "system: IPv4 Works",
          "slug": "system-ipv4-works"
        },
        {
          "name": "system: Resolves AAAA Correctly",
          "slug": "system-resolves-aaaa-correctly"
        },
        {
          "name": "system: IPv6 Works",
          "slug": "system-ipv6-works"
        },
        {
          "name": "system: Resolves A Correctly",
          "slug": "system-resolves-a-correctly"
        },
        {
          "name": "system: IPv6 Capable",
          "slug": "system-ipv6-capable"
        },
        {
          "name": "system: IPv6 Stable 90d",
          "slug": "system-ipv6-stable-90d"
        },
        {
          "name": "system: IPv4 Capable",
          "slug": "system-ipv4-capable"
        },
        {
          "name": "system: IPv4 Stable 1d",
          "slug": "system-ipv4-stable-1d"
        }
      ],
      "total_uptime": 25761608,
      "type": "Probe"
    }
  ]
}
```
