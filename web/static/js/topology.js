/**
 * Created by Xie on 23/2/16.
 */
//depend on d3.js
;
(function () {
    window.topology = {} || topology;

    var circle_r = 15;
    topology.display = function () {
        topology.init_promise = $.ajax("/api/v1/dbdata/topology", {
            method: "GET"
        });
        topology.init_promise.then(function (d, d2,data3) {
            console.log(data3.getAllResponseHeaders());
            topology.data_json = JSON.parse(d);
            console.log(topology.data_json);
            topology.links = topology.data_json.data.links;
            topology.nodes = topology.data_json.data.nodes;
            topology.nodes_map = {};
            topology.nodes.forEach(function (node) {
                topology.nodes_map[node.id] = node;
            });
            topology.links.forEach(function (link) {
                link.target = topology.nodes_map[link.targethostid];
                link.source = topology.nodes_map[link.sourcehostid];
            });


            topology.force = d3.layout.force()
                .nodes(topology.nodes)
                .links(topology.links)
                .size([$("#container").width(), window.innerHeight - 200])
                .linkDistance(160)
                .charge(-1000)
                .on("tick", tick)
                .start();

            if (topology.path) {
                topology.path.remove();
            }
            topology.path = topology.svg.append("g").selectAll("path")
                .data(topology.force.links())
                .enter().append("path")
                .attr("class", function (d) {
                    return "link " + d.type;
                })
                .attr("marker-end", function (d) {
                    return "url(#" + d.type + ")";
                });


            if (topology.circle) {
                topology.circle.remove();
            }
            topology.circle = topology.svg.append("g").selectAll("circle")
                .data(topology.force.nodes())
                .enter().append("circle")
                .attr("data-id", function (d) {
                    return d.id;
                })
                .attr("r", circle_r)
                .attr("class", "database")
                .attr("onclick", "topology.onNodeClick(this);")
                .attr("onmouseover", "topology.onMouseover(this, window.event);")
                .attr("onmouseout", "topology.onMouseout(this);")
                .call(topology.force.drag);

            if (topology.text) {
                topology.text.remove();
            }
            topology.text = topology.svg.append("g").selectAll("text")
                .data(topology.force.nodes())
                .enter().append("text")
                .attr("x", 8)
                .attr("y", ".31em")
                .attr("class", "database")
                .text(function (d) {
                    return d.name;
                });
        });
    };

    topology.display();

    topology.svg = d3.select("#container").append("svg")
        .attr("class", "topology");

    // Per-type markers, as they don't inherit styles.
    topology.svg.append("defs").selectAll("marker")
        .data(["connect", "disconnect"])
        .enter().append("marker")
        .attr("id", function (d) {
            return d;
        })
        .attr("class", function (d) {
            return "marker " + d;
        })
        .attr("viewBox", "0 -5 10 10")
        .attr("refX", 20)
        .attr("refY", -1)
        .attr("markerWidth", 10)
        .attr("markerHeight", 10)
        .attr("orient", "auto")
        .append("path")
        .attr("d", "M0,-5L10,0L0,5Z");

    d3.select("#container")
        .append(function () {
            var div = document.createElement("div");
            div.setAttribute("id", "db_info");
            div.setAttribute("class", "info-board");
            return div;
        });

    topology.onNodeClick = function (target) {

    };

    function json2str (js) {
        var info = JSON.stringify(js).split(/,|{|}/);
        info.pop();
        info.shift();
        info = info.join('\n');
        return info;
    }

    topology.onMouseover = function (target, event) {
        //var node_info = topology.nodes_map[$(target).data('id')];
        //var float_div = $("#db_info");
        //var x = event.clientX + 10 + "px";
        //var y = event.clientY + 10 + "px";
        //float_div[0].innerText = json2str(node_info);
        //float_div.css("left", x);
        //float_div.css("top", y);
        //float_div.slideDown();
        target.style.cursor = "hand";
    };

    topology.onMouseout = function (target) {
        //$("#db_info").slideUp();
        target.style.cursor = "default";
    };

    function tick() {
        topology.path.attr("d", linkArc);
        topology.circle.attr("transform", transform);
        topology.text.attr("transform", transform);
    }

    function linkArc(d) {
        var dx = d.target.x - d.source.x,
            dy = d.target.y - d.source.y,
            dr = Math.sqrt(dx * dx + dy * dy);
        return "M" + d.source.x + "," + d.source.y + "A" + dr + "," + dr + " 0 0,1 " + d.target.x + "," + d.target.y;
    }

    function transform(d) {
        return "translate(" + d.x + "," + d.y + ")";
    }
})();
